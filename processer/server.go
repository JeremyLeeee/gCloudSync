package processer

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/metadata"
	"gcloudsync/network"
	"log"
)

// main entry
func StartServer() {
	server := network.NewServer(config.Port)

	done := make(chan bool)
	go server.Listen()
	go handleServer(server, done)

	log.Println("start listening")
	<-done
}

// enter main loop for data exchanging
func handleServer(server network.ITCPServer, done chan bool) {
	bc := server.GetBuffChan()
	var buffer []byte
	for {
		data := <-bc
		log.Println("--received--:", string(data))

		tag, length, err := metadata.GetHeaderFromData(data)
		common.ErrorHandleDebug(err)
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

		log.Println("tag:", tag, "len:", length, "data:", string(currentData))
		// now buffer contains all the data
		switch tag {
		case common.SysInit:
			log.Println("client initing...")
			sendServer(server, common.SysDone, []byte{})
		default:
			log.Panic("unknown op")
		}
		// clear current data buffer
		buffer = buffer[length+24:]
		// log.Println("remain buffer:", string(buffer))
	}
}

// add header and send
func sendServer(server network.ITCPServer, op common.SysOp, data []byte) error {
	// get header
	header := metadata.NewHeader(uint64(len(data)), op)
	sendByte, err := header.ToByteArray()
	common.ErrorHandleDebug(err)

	// merge to array
	sendByte = common.MergeArray(sendByte, data)

	log.Println("send:", string(sendByte), "len:", len(sendByte))
	server.Send(sendByte)

	return err
}
