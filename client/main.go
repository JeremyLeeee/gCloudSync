package main

import (
	"fmt"
	"gcloudsync/config"
	"gcloudsync/fswatcher"
	"gcloudsync/network"
	"log"
	"os"
)

func main() {
	testNetwork()
}

func testNetwork() {
	client := network.NewClient(config.ServerIP, config.Port)
	client.Connect()
	log.Println("connect ok")

	var line string
	for {
		fmt.Scanln(&line)
		client.Send([]byte(line))
		log.Println("write: " + line)
		recvBuffer := client.Receive()
		log.Println("receive: ", string(recvBuffer))
	}
}

func testFsWatcher() {
	path := os.Args[1]
	fw := fswatcher.NewFsWatcher(path)
	fschan := fw.GetChan()

	// start watching
	go func() {
		fw.StartWatching()
	}()
	log.Println("start watching folder:", path)

	// dispatch fs event
	for {
		event := <-fschan
		log.Print(event)
	}
}
