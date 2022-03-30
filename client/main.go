package main

import (
	"fmt"
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fswatcher"
	"gcloudsync/metadata"
	"gcloudsync/network"
	"log"
	"os"
)

func main() {
	header := metadata.NewHeader(12, common.SysDone)
	buf, err := header.ToByteArray()
	common.ErrorHandleDebug(err)
	log.Println(buf)

	tag, datalen, err := header.GetHeaderFromData(buf)
	common.ErrorHandleDebug(err)

	log.Println("tag:", tag, "datalen:", datalen)

	//testNetwork()
	//testFsWatcher()
}

func HandleClient(client network.ITCPClient) {
	bc := client.GetBuffChan()
	var buffer []byte
	for {
		buffer = <-bc
		log.Println(string(buffer))
		// client.Send([]byte("The data to Server"))
	}
}

func testNetwork() {
	client := network.NewClient(config.ServerIP, config.Port)
	client.Connect()
	log.Println("connect ok")

	go HandleClient(client)
	go client.ReadFromClient()

	var line string
	for {
		fmt.Scanln(&line)
		log.Println([]byte(line))
		client.Send([]byte(line))
	}
}

func testFsWatcher() {
	path := os.Args[1]
	fw := fswatcher.NewFsWatcher(path)
	fschan := fw.GetChan()

	// start watching
	go fw.StartWatching()

	log.Println("start watching folder:", path)

	// dispatch fs event
	for {
		event := <-fschan
		log.Print(event)
	}
}
