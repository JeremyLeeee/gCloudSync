package core

import (
	"gcloudsync/internal/config"
	"gcloudsync/internal/network"
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
	bc := s.server.GetBuffChan()
	go s.server.Listen()
	go handleCore(s.server, bc, done, nil, nil)

	log.Println(logtag, "start listening")
	<-done
}
