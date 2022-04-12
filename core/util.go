package core

import (
	"errors"
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
	"gcloudsync/metadata"
	"log"
	"reflect"
)

var logtag string = "[Core]"

func wrappAndSend(base interface{}, op common.SysOp, data []byte, last uint32) error {
	// get header
	header := metadata.NewHeader(uint32(len(data)), op, last)
	sendByte, err := header.ToByteArray()
	common.ErrorHandleDebug(logtag, err)

	// merge to array
	sendByte = common.MergeArray(sendByte, data)

	log.Println(logtag, "send:", string(data), "len:", len(data))

	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(sendByte)
	reflect.ValueOf(base).MethodByName("Send").Call(in)

	return err
}

// main loog for data exchanging
// @base: interface for server or client
// @bufferChan: buffer channel for comming data
// @done: a bool channel represent whether everything is done
func handleCore(base interface{}, bufferChan chan []byte, done chan bool) {
	var buffer []byte
	var header metadata.Header
	var data []byte
	var err error
	var isClient bool

	baseTypeString := reflect.TypeOf(base).String()
	log.Println(logtag, "current base:", baseTypeString)

	if baseTypeString == "*network.TCPClient" {
		isClient = true
	} else {
		isClient = false
	}

	// main loop for data processing
	for {
		tempBuffer := <-bufferChan
		buffer = common.MergeArray(buffer, tempBuffer)
		// log.Println(logtag, "received:", string(data))

		for {
			buffer, header, data, err = getOnePackageFromBuffer(buffer)
			if err != nil {
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
					syncOneFileSend(fsops.RemoveRootPrefix(filePath), base, isClient)
				}
				wrappAndSend(base, common.SysDone, []byte{}, common.IsLastPackage)
			case common.SysSyncFileEmpty:
				// transfer the file directly
			case common.SysSyncFileNotEmpty:
				// apply rsync algo
			case common.SysSyncFolder:
				// generate new folder
				folderPath := string(data)

				if isClient {
					err = fsops.Makedir(config.ClientRootPath + folderPath)
				} else {
					err = fsops.Makedir(config.ServerRootPath + folderPath)
				}
				common.ErrorHandleDebug(logtag, err)
			case common.SysSyncFileBegin:
				// entry for transfering file
				log.Println(logtag, "file to be transfered:", string(data))
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

	if isClient {
		ok, _ = fsops.IsFolder(config.ClientRootPath + path)
	} else {
		ok, _ = fsops.IsFolder(config.ServerRootPath + path)
	}

	if ok {
		// a folder
		wrappAndSend(base, common.SysSyncFolder, []byte(path), common.IsLastPackage)
	} else {
		// a file
		wrappAndSend(base, common.SysSyncFileBegin, []byte(path), common.IsLastPackage)
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
