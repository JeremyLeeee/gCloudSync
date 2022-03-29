package main

import (
	"gcloudsync/config"
	"gcloudsync/network"
	"log"
)

func main() {
	server := network.NewServer(config.Port)

	done := make(chan bool)
	go server.Listen()
	go HandleServer(server)

	log.Println("start listening")
	<-done
}

// receive data from client and response
func HandleServer(server network.ITCPServer) {
	bc := server.GetBuffChan()
	var buffer []byte
	for {
		buffer = <-bc
		log.Println(string(buffer))
		server.Send([]byte("The data to Client"))
	}
}
