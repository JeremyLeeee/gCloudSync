package core

import (
	"gcloudsync/internal/common"
	"gcloudsync/internal/config"
	"gcloudsync/internal/fsops"
	"gcloudsync/internal/fswatcher"
	"gcloudsync/internal/network"
	"log"
	"strings"
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

	log.Println(logtag, "connected successfully.")
	done := make(chan bool)
	initDone := make(chan bool)
	bc := c.client.GetBuffChan()

	// handle received message
	go handleCore(c.client, bc, done, c.eventChan, c.eventDone)

	// start receiving
	go c.client.ReadFromServer()

	// init config

	c.syncConfig()
	log.Println(logtag, "sync config...")

	// init config ok
	<-done

	log.Println(logtag, "sync config ok.")

	// send init signal
	WrappAndSend(c.client, common.SysInit, []byte{}, common.IsLastPackage)
	log.Println(logtag, "sync all files...")

	// init file list ok
	<-done

	go c.startEventLoop(c.eventDone, initDone)
	// process the remain event in eventChen
	// before start watching fs
	<-initDone
	// next start watching fs

	go c.startWatching()

	<-done
	close(done)
}

func (c *ClientCore) syncConfig() {
	data := config.GetConfig().ToBytes()
	WrappAndSend(c.client, common.SysInitSyncConfig, data, common.IsLastPackage)
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
		// log.Println(logtag, event)
		c.eventChan <- event
	}
}

func (c *ClientCore) startEventLoop(eventDone chan bool, initDone chan bool) {
	log.Println(logtag, "start event loop...")
	inited := false
	if len(c.eventChan) == 0 && !inited {
		initDone <- true
		inited = true
		WrappAndSend(c.client, common.SysInitFinished, []byte{}, common.IsLastPackage)
	}
	for {
		if len(c.eventChan) == 0 && !inited {
			initDone <- true
			inited = true
			WrappAndSend(c.client, common.SysInitFinished, []byte{}, common.IsLastPackage)
		}
		event := <-c.eventChan
		// log.Println(logtag, "process event:", event)
		currentFilePath = event.FileName
		path := fsops.RemoveRootPrefix(event.FileName, true)

		// emit macos .DS_Store file
		if strings.Compare(event.FileName[len(event.FileName)-8:], "DS_Store") == 0 {
			continue
		}
		switch event.Op {
		case common.OpFetch:
			// sync file
			if fsops.IsFileExist(event.FileName) {
				// rsync
				// log.Println(logtag, "need rsync")
				// get md5
				checksum := common.GetFileMd5(event.FileName)
				// package structure:
				// +------------+--------------------+
				// |checksum    |filename            |
				// +------------+--------------------+
				// <---16bytes-->

				data := common.MergeArray(checksum, []byte(path))
				// log.Println(logtag, "sync:", event.FileName)
				WrappAndSend(c.client, common.SysSyncFileNotEmpty, data, common.IsLastPackage)
			} else {
				// direct file
				log.Println(logtag, "fetch:", event.FileName)
				WrappAndSend(c.client, common.SysSyncFileEmpty, []byte(path), common.IsLastPackage)
			}
		case common.OpCreate:
			if inited {
				log.Println(logtag, "create:", event.FileName)
			}
			WrappAndSend(c.client, common.SysOpCreate, []byte(path), common.IsLastPackage)
		case common.OpModify:
			if inited {
				log.Println(logtag, "modify:", event.FileName)
			}
			WrappAndSend(c.client, common.SysOpModify, []byte(path), common.IsLastPackage)
		case common.OpRename:
			if inited {
				log.Println(logtag, "rename from:", event.OriginFile)
				log.Println(logtag, "to:", event.FileName)
			}
			data := RenameEventToBytes(event)
			WrappAndSend(c.client, common.SysOpRename, []byte(data), common.IsLastPackage)
		case common.OpRemove:
			if inited {
				log.Println(logtag, "remove:", event.FileName)
			}
			WrappAndSend(c.client, common.SysOpRemove, []byte(path), common.IsLastPackage)
		case common.OpMkdir:
			if inited {
				log.Println(logtag, "mkdir:", event.FileName)
			}
			WrappAndSend(c.client, common.SysOpMkdir, []byte(path), common.IsLastPackage)
		case common.OpChmod:
			// do nothing
			continue
		default:
			log.Panic(logtag, "unknown event")
		}

		// if handleCore finished current event
		// eventDone will be released
		<-eventDone
	}
}
