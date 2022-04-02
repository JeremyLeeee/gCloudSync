package core

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fswatcher"
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
	bc := c.client.GetBuffChan()

	// handle received message
	go handleCore(c.client, bc, done)

	// start receiving
	go c.client.ReadFromServer()

	// send init signal
	wrappAndSend(c.client, common.SysInit, []byte{}, common.IsLastPackage)

	// init ok
	<-done

	// next start watching fs
	// fall into another loop
	c.startWatching()

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
	}
}
