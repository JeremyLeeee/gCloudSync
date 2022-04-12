package core

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
	"gcloudsync/fswatcher"
	"gcloudsync/network"
	"log"
	"time"
)

type ClientCore struct {
	client    network.ITCPClient
	watchPath string
	eventChan chan common.FsEvent
	eventDone chan bool
}

func NewClientCore(path string) ClientCore {
	eventChan := make(chan common.FsEvent, config.EventChanSize)
	eventDone := make(chan bool)
	cli := network.NewClient(config.ServerIP, config.Port)
	return ClientCore{client: cli, watchPath: path,
		eventChan: eventChan, eventDone: eventDone}
}

func (c *ClientCore) StartClient() {
	err := c.client.Connect()
	common.ErrorHandleFatal(logtag, err)
	defer c.client.Close()

	done := make(chan bool)
	initDone := make(chan bool)
	bc := c.client.GetBuffChan()

	// handle received message
	go handleCore(c.client, bc, done, c.eventChan, c.eventDone)

	// start receiving
	go c.client.ReadFromServer()

	// send init signal
	WrappAndSend(c.client, common.SysInit, []byte{}, common.IsLastPackage)
	log.Println(logtag, "initializing...")

	// init ok
	<-done

	go c.startEventLoop(c.eventDone, initDone)
	// process the remain event in eventChen
	// before start watching fs
	<-initDone
	// next start watching fs
	// fall into another loop

	// in case write file trigger watcher
	log.Println(logtag, "init fishied, wait for flushing...")
	time.Sleep(time.Second * 2)

	go c.startWatching()

	<-done
	close(done)
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
		c.eventChan <- event
	}
}

func (c *ClientCore) startEventLoop(eventDone chan bool, initDone chan bool) {
	log.Println(logtag, "start event loop...")
	for {
		event := <-c.eventChan
		if len(c.eventChan) == 0 {
			initDone <- true
		}
		// log.Println(logtag, "process event:", event)
		currentFilePath = event.FileName
		switch event.Op {
		case common.OpFetch:
			// sync file
			if fsops.IsFileExist(event.FileName) {
				// rsync
				// log.Println(logtag, "need rsync")
				continue
			} else {
				// direct file
				log.Println(logtag, "fetch:", event.FileName)
				WrappAndSend(c.client, common.SysSyncFileEmpty,
					[]byte(fsops.RemoveRootPrefix(event.FileName, true)),
					common.IsLastPackage)
			}
		}

		// if handleCore finished current event
		// eventDone will be released
		<-eventDone
	}
}
