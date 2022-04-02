package core

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fswatcher"
	"gcloudsync/metadata"
	"gcloudsync/network"
	"log"
)

var logtag string = "[Core]"

type ClientCore struct {
	client    network.ITCPClient
	watchPath string
}

func NewClientCore(path string) ClientCore {
	cli := network.NewClient(config.ServerIP, config.Port)
	return ClientCore{client: cli, watchPath: path}
}

func (c *ClientCore) StartClient() {
	err := c.client.Connect()
	common.ErrorHandleFatal(logtag, err)
	defer c.client.Close()

	done := make(chan bool)

	// handle received message
	go c.handleClient(done)

	// start receiving
	go c.client.ReadFromClient()

	// send init signal
	WrappAndSend(c.client, common.SysInit, []byte{}, common.IsLastPackage)

	// init ok
	<-done

	// next start watching fs
	// fall into another loop
	c.startWatching()

	close(done)
}

// enter main loop for data exchanging
func (c *ClientCore) handleClient(done chan bool) {
	bc := c.client.GetBuffChan()
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
		case common.SysDone:
			log.Println(logtag, "Done")
			done <- true
		default:
			log.Panic(logtag, "unknown op")

		}
		// clear current data buffer
		buffer = buffer[length+24:]
		// log.Println("remain buffer:", string(buffer))
	}
}

// start watching fs
func (c *ClientCore) startWatching() {
	fw := fswatcher.NewFsWatcher(c.watchPath)
	fschan := fw.GetChan()

	// start watching
	go fw.StartWatching()

	log.Println(logtag, "start watching folder:", c.watchPath)

	// dispatch fs event
	for {
		event := <-fschan
		log.Println(logtag, event)
	}
}
