package core

import (
	"bytes"
	"errors"
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
	"gcloudsync/metadata"
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
	log.Println(logtag, "current base:", baseTypeString)

	// store file data
	databuff := make([]byte, config.MaxBufferSize)

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
				WrappAndSend(base, common.SysDone, []byte{}, common.IsLastPackage)

			case common.SysSyncFileEmpty:
				// transfer the file directly
				path := string(data)
				absPath := pathPrefix + path
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
			case common.SysSyncFileNotEmpty:
				// receive checksum from sender
				checksum := data[0:16]
				absPath := pathPrefix + string(data[16:])

				// validate local file
				md5 := common.GetFileMd5(absPath)
				if bytes.Equal(md5, checksum) {
					// no need to sync
					// log.Println(logtag, absPath, "no need to sync")
					WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)
				} else {
					// apply rsync algo
					log.Println(logtag, absPath, "need rsync")
					WrappAndSend(base, common.SysSyncFinished, []byte{}, common.IsLastPackage)
				}

			case common.SysSyncFolder:
				// generate new folder
				folderPath := string(data)
				err = fsops.Makedir(pathPrefix + folderPath)
				common.ErrorHandleDebug(logtag, err)

			case common.SysSyncFileBegin:
				// entry for transfering file
				// log.Println(logtag, "file to be transfered:", string(data))
				path := pathPrefix + string(data)

				// add to event loop
				if eventChan != nil {
					event := common.FsEvent{Op: common.OpFetch, FileName: path, IsDir: false}
					eventChan <- event
				}
			case common.SysSyncFileDirect:
				// log.Println(logtag, "write file:", currentFilePath, "datalen:", len(data))
				_, err := fsops.Write(currentFilePath, data, 0)
				common.ErrorHandleDebug(logtag, err)
				if header.Last == common.IsLastPackage {
					fsops.CloseCurrentFile()
					if eventChan != nil {
						eventDone <- true
					}
				}
			case common.SysSyncFinished:
				// log.Println(logtag, "finish syncing:", currentFilePath)
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
		WrappAndSend(base, common.SysSyncFolder, []byte(path), common.IsLastPackage)
	} else {
		WrappAndSend(base, common.SysSyncFileBegin, []byte(path), common.IsLastPackage)
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
