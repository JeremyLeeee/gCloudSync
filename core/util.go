package core

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
	"gcloudsync/metadata"
	"log"
	"reflect"
)

func wrappAndSend(base interface{}, op common.SysOp, data []byte, last uint32) error {
	// get header
	header := metadata.NewHeader(uint32(len(data)), op, last)
	sendByte, err := header.ToByteArray()
	common.ErrorHandleDebug(logtag, err)

	// merge to array
	sendByte = common.MergeArray(sendByte, data)

	log.Println(logtag, "send:", string(sendByte), "len:", len(sendByte))

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
	for {
		data := <-bufferChan
		log.Println(logtag, "received:", string(data))

		tag, length, err := metadata.GetHeaderFromData(data)
		common.ErrorHandleDebug(logtag, err)
		gotHeader := false

		if err != nil {
			if err.Error() == "invalid length" && !gotHeader {
				// smaller than header
				// first get metadata from header to get correct length
				buffer = common.MergeArray(buffer, data)
				tag, length, err = metadata.GetHeaderFromData(buffer)
				if err != nil {
					gotHeader = true
					continue
				}
			} else if err.Error() == "invalid signature" {
				buffer = common.MergeArray(buffer, data)
				continue
			}
		}

		if len(buffer) == 0 {
			// successfully get all valid data for the first time
			buffer = common.MergeArray(buffer, data)
		}

		// expect more
		if int(length) > len(buffer)-24 {
			continue
		}

		// get current data buffer
		currentData := buffer[0 : length+24]

		log.Println(logtag, "tag:", tag, "len:", length, "data:", string(currentData))

		// processing different system event
		switch tag {
		case common.SysInit:
			// server respond client init
			log.Println(logtag, "client initing...")
			// get all file list and send to client
			flist := fsops.GetAllFile(config.ServerRootPath)
			common.ErrorHandleDebug(logtag, err)
			// for each file and folder, sync to client
			for _, filePath := range flist {
				log.Println(logtag, "Files in Server:", fsops.RemoveRootPrefix(filePath))
				// syncOneFile(fsops.RemoveRootPrefix(filePath), base)
			}
			wrappAndSend(base, common.SysDone, []byte{}, common.IsLastPackage)
		case common.SysSyncFileEmpty:
			// transfer the file directly
		case common.SysSyncFileNotEmpty:
			// apply rsync algo
		case common.SysDone:
			done <- true
		default:
			log.Panic(logtag, "unknown op")
		}
		// clear current data buffer
		buffer = buffer[length+24:]
		// log.Println("remain buffer:", string(buffer))
	}

}

// sync one file with peer
// @path: relative path of the file or folder
func syncOneFile(path string, base interface{}) {
	ok, err := fsops.IsFolder(path)
	common.ErrorHandleDebug(logtag, err)
	if ok {
		// a folder
		wrappAndSend(base, common.SysSyncFolder, []byte(path), common.IsLastPackage)
	} else {
		// a file
		wrappAndSend(base, common.SysSyncFileBegin, []byte(path), common.IsLastPackage)
	}
}
