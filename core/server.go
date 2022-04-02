package core

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
	"gcloudsync/metadata"
	"gcloudsync/network"
	"log"
)

type ServerCore struct {
	server network.ITCPServer
	path   string
}

func NewServerCore(path string) ServerCore {
	server := network.NewServer(config.Port)
	return ServerCore{server: server, path: path}
}

// main entry
func (s *ServerCore) StartServer() {
	done := make(chan bool)
	go s.server.Listen()
	go s.handleServer(done)

	log.Println(logtag, "start listening")
	<-done
}

// enter main loop for data exchanging
func (s *ServerCore) handleServer(done chan bool) {
	bc := s.server.GetBuffChan()
	var buffer []byte
	for {
		data := <-bc
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
		// now buffer contains all the data
		switch tag {
		case common.SysInit:
			log.Println(logtag, "client initing...")
			// get all file list and send to client
			flist := fsops.GetAllFile(s.path)
			common.ErrorHandleDebug(logtag, err)
			// for each file and folder, sync to client
			for _, filePath := range flist {
				log.Println(logtag, "[Files in Server]", fsops.RemoveRootPrefix(filePath))
			}
			WrappAndSend(s.server, common.SysDone, []byte{}, common.IsLastPackage)
		default:
			log.Panic(logtag, "unknown op")
		}
		// clear current data buffer
		buffer = buffer[length+24:]
		// log.Println("remain buffer:", string(buffer))
	}
}
