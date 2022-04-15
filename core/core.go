package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
	"gcloudsync/metadata"
	"gcloudsync/rsync"
	"log"
	"reflect"
)

var logtag string = "[Core]"
var currentFilePath string

func WrappAndSend(base interface{}, op common.SysOp, data []byte, last uint32) error {
	// get header
	header := metadata.NewHeader(uint32(len(data)), op, last)
	sendByte, err := header.ToByteArray()
	common.ErrorHandleDebug(logtag, err)

	// merge to array
	sendByte = common.MergeArray(sendByte, data)

	// log.Println(logtag, "send:", string(data), "len:", len(data))

	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(sendByte)
	reflect.ValueOf(base).MethodByName("Send").Call(in)

	return err
}

// main loog for data exchanging
// @base: interface for server or client
// @bufferChan: buffer channel for comming data
// @done: a bool channel represent whether everything is done
func handleCore(base interface{}, bufferChan chan []byte, done chan bool,
	eventChan chan common.FsEvent, eventDone chan bool) {

	var buffer []byte
	var header metadata.Header
	var data []byte
	var err error
	var isClient bool
	var pathPrefix string

	baseTypeString := reflect.TypeOf(base).String()
	// log.Println(logtag, "current base:", baseTypeString)

	if baseTypeString == "*network.TCPClient" {
		isClient = true
		pathPrefix = config.ClientRootPath
	} else {
		isClient = false
		pathPrefix = config.ServerRootPath
	}

	// main loop for data processing
	for {
		tempBuffer := <-bufferChan
		buffer = common.MergeArray(buffer, tempBuffer)

		// log.Println(logtag, "bufferlen:", len(tempBuffer))
		for {
			buffer, header, data, err = getOnePackageFromBuffer(buffer)
			if err != nil {
				// log.Println(logtag, "invalid package")
				break
			}
			// processing different system event
			// log.Println(logtag, "event tag:", header.Tag)
			switch header.Tag {
			case common.SysInit:
				// server respond client init
				log.Println(logtag, "client initing...")
				// get all file list and send to client
				flist := fsops.GetAllFile(config.ServerRootPath)
				common.ErrorHandleDebug(logtag, err)
				// for each file and folder, sync to client
				for _, filePath := range flist {
					syncOneFileSend(fsops.RemoveRootPrefix(filePath, false), base, isClient)
				}
				WrappAndSend(base, common.SysInitUpload, []byte{}, common.IsLastPackage)

			case common.SysInitUpload:
				// for files not exist in server
				// upload to keep in consistance
				log.Println(logtag, "upload new files...")
				flist := fsops.GetAllFile(config.ClientRootPath)
				common.ErrorHandleDebug(logtag, err)
				var op common.FsOp
				// add to event loop
				for _, filePath := range flist {
					if ok, _ := fsops.IsFolder(filePath); ok {
						op = common.OpMkdir
					} else {
						op = common.OpCreate
					}
					event := common.FsEvent{Op: op, FileName: filePath}
					eventChan <- event
				}

				done <- true

			case common.SysInitSyncConfig:
				cg := config.GetConfig()
				cg.ConfigFromBytes(data)
				log.Println(logtag, "config sync finished.")

				config.PrintCurrentConfig()
				WrappAndSend(base, common.SysDone, []byte{}, common.IsLastPackage)

			case common.SysInitSyncFolder:
				absPath := pathPrefix + string(data)
				if !fsops.IsFileExist(absPath) {
					fsops.Makedir(absPath)
				}

			case common.SysInitSyncFile:
				// entry for transfering file
				// log.Println(logtag, "file to be transfered:", string(data))
				path := pathPrefix + string(data)

				// add to event loop
				if eventChan != nil {
					event := common.FsEvent{Op: common.OpFetch, FileName: path}
					eventChan <- event
				}

			case common.SysInitFinished:
				// only server will receive this
				log.Println(logtag, "client init finished.")

			case common.SysSyncFileEmpty:
				// transfer the file directly
				path := string(data)
				absPath := pathPrefix + path
				directFileSend(base, absPath)

			case common.SysSyncFileNotEmpty:
				// receive checksum from sender
				checksum := data[0:16]
				path := string(data[16:])
				absPath := pathPrefix + path

				// validate local file
				md5 := common.GetFileMd5(absPath)
				if bytes.Equal(md5, checksum) {
					// no need to sync
					// log.Println(logtag, absPath, "no need to sync")
					WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)
				} else {
					// apply rsync algo
					log.Println(logtag, absPath, "need rsync")
					currentFilePath = absPath
					WrappAndSend(base, common.SysOpModify, []byte(path), common.IsLastPackage)
				}

			case common.SysSyncFileDirect:
				// log.Println(logtag, "write file:", currentFilePath, "datalen:", len(data))
				_, err := fsops.Write(currentFilePath, data, 0)
				common.ErrorHandleDebug(logtag, err)
				if header.Last == common.IsLastPackage {
					fsops.CloseCurrentFile()
					if eventChan != nil {
						eventDone <- true
					} else {
						// server receive file
						WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)
					}
				}
			case common.SysSyncFinished:
				log.Println(logtag, "sync finished:", currentFilePath)
				if eventDone != nil {
					eventDone <- true
				}

			case common.SysOpCreate:
				absPath := pathPrefix + string(data)
				// 1. touch
				err := fsops.Create(absPath)
				log.Println(logtag, "create:", absPath)
				common.ErrorHandleDebug(logtag, err)
				// 2. sync
				currentFilePath = absPath
				WrappAndSend(base, common.SysSyncFileEmpty, []byte(string(data)), common.IsLastPackage)

			case common.SysOpRemove:
				absPath := pathPrefix + string(data)
				err := fsops.Delete(absPath)
				log.Println(logtag, "remove:", absPath)
				common.ErrorHandleDebug(logtag, err)
				WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)

			case common.SysOpMkdir:
				// generate new folder
				absPath := pathPrefix + string(data)
				err = fsops.Makedir(absPath)
				log.Println(logtag, "mkdir:", absPath)
				common.ErrorHandleDebug(logtag, err)
				WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)
			case common.SysOpRename:
				event := BytesToRenameEvent(data)
				new := pathPrefix + event.FileName
				old := pathPrefix + event.OriginFile
				log.Println(logtag, "rename from:", old)
				log.Println(logtag, "to:", new)
				err := fsops.Rename(old, new)
				common.ErrorHandleDebug(logtag, err)
				WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)

			case common.SysOpModify:
				// both client and server can get here
				// generate checksum
				path := string(data)
				absPath := pathPrefix + path

				// if file not exist, create one
				if fsops.IsFileExist(absPath) {
					fsops.Create(absPath)
				}

				currentFilePath = absPath
				cks := rsync.GetCheckSums(absPath)

				log.Println(logtag, "modifying:", absPath)
				WrappAndSend(base, common.SysSyncGenerateDiff, cks, common.IsLastPackage)

			case common.SysSyncGenerateDiff:
				diff, err := rsync.GetDiff(data, currentFilePath)
				common.ErrorHandleDebug(logtag, err)

				WrappAndSend(base, common.SysSyncReformFile, diff, common.IsLastPackage)

			case common.SysSyncReformFile:

				err := rsync.ReformFile(data, currentFilePath)
				log.Println(logtag, "sync finished:", currentFilePath)
				common.ErrorHandleDebug(logtag, err)

				WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)

				if eventDone != nil {
					eventDone <- true
				}

			case common.SysDone:
				done <- true
			default:
				log.Panic(logtag, "unknown op")
			}
		}
	}

}

// sync one file with peer
// @path: relative path of the file or folder
func syncOneFileSend(path string, base interface{}, isClient bool) {
	var ok bool
	var absPath string
	if isClient {
		absPath = config.ClientRootPath + path
	} else {
		absPath = config.ServerRootPath + path
	}

	ok, _ = fsops.IsFolder(absPath)

	if ok {
		WrappAndSend(base, common.SysInitSyncFolder, []byte(path), common.IsLastPackage)
	} else {
		WrappAndSend(base, common.SysInitSyncFile, []byte(path), common.IsLastPackage)
	}
}

func getOnePackageFromBuffer(buffer []byte) (remainBuffer []byte, header metadata.Header, packageData []byte, err error) {
	header, err = metadata.GetHeaderFromData(buffer)
	// common.ErrorHandleDebug(logtag, err)
	// log.Println(logtag, "buffer len:", len(buffer))

	// the error case include:
	if err != nil {
		// 1. data buffer size smaller than header
		// 2. invalid signiture
		return buffer, header, nil, err
	}

	if len(buffer) < int(header.Length)+24 {
		// 3. expect more data
		return buffer, header, nil, errors.New("expect more")
	}

	packageData = buffer[24 : header.Length+24]
	remainBuffer = buffer[header.Length+24:]

	return remainBuffer, header, packageData, nil
}

func directFileSend(base interface{}, absPath string) {
	databuff := make([]byte, config.MaxBufferSize)

	fileSize, err := fsops.GetFileSize(absPath)
	common.ErrorHandleDebug(logtag, err)

	// start sending file
	var count int64
	count = 0

	// log.Println(logtag, "sync:", path)
	for {
		n, _ := fsops.ReadOnce(absPath, databuff, count)
		count = count + int64(n)

		filedata := databuff[0:n]
		if n == int(fileSize) {
			// read once is enough
			WrappAndSend(base, common.SysSyncFileDirect, filedata, common.IsLastPackage)
			break
		} else {
			WrappAndSend(base, common.SysSyncFileDirect, filedata, common.IsNotLastPacage)
			if count == fileSize {
				log.Println(logtag, absPath, "read finished")
				WrappAndSend(base, common.SysSyncFileDirect, []byte{}, common.IsLastPackage)
				break
			}
		}
	}
	fsops.CloseCurrentFile()
}

func RenameEventToBytes(fe common.FsEvent) (b []byte) {
	if fe.Op == common.OpRename {
		// package structure:
		// +-----+----------+-----+----------+
		// | len | new path | len | old path |
		// +-----+----------+-----+----------+
		// |  4  |          |  4  |          |
		// +-----+----------+-----+----------+
		var buf bytes.Buffer
		b := make([]byte, 4)

		new := fsops.RemoveRootPrefix(fe.FileName, true)
		old := fsops.RemoveRootPrefix(fe.OriginFile, true)
		binary.BigEndian.PutUint32(b, uint32(len(new)))
		buf.Write([]byte(b))
		buf.Write([]byte(new))
		binary.BigEndian.PutUint32(b, uint32(len(old)))
		buf.Write([]byte(b))
		buf.Write([]byte(old))

		return buf.Bytes()
	}
	return b
}

func BytesToRenameEvent(b []byte) (fe common.FsEvent) {

	newlen := int(binary.BigEndian.Uint32(b[0:4]))
	new := string(b[4 : 4+newlen])
	oldlen := int(binary.BigEndian.Uint32(b[4+newlen : 8+newlen]))
	old := string(b[8+newlen : 8+newlen+oldlen])

	fe.FileName = new
	fe.OriginFile = old
	fe.Op = common.OpRename

	return
}
