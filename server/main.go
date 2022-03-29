package main

import (
	"gcloudsync/config"
	"gcloudsync/network"
	"log"
)

// receive data from client and response
func HandleServer(server network.ITCPServer) {
	bc := server.GetBuffChan()
	for {
		buffer := <-bc
		log.Println(string(buffer))
		server.Send([]byte("welcome"))
	}
}

func main() {
	server := network.NewServer(config.Port)

	done := make(chan bool)
	go server.Listen()
	go HandleServer(server)

	log.Println("start listening")
	<-done
}
