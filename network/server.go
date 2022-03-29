package network

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"log"
	"net"
)

type ITCPServer interface {
	Listen()
	GetBuffChan() chan []byte
	Send(b []byte) error
	readFromClient()
}

type TCPServer struct {
	port        string
	buffchan    chan []byte
	currentConn net.Conn
}

func NewServer(port string) ITCPServer {
	buffchan := make(chan []byte)
	return &TCPServer{port: port, buffchan: buffchan}
}

func (s *TCPServer) Listen() {
	tcpServer, _ := net.ResolveTCPAddr("tcp4", ":"+config.Port)
	listener, _ := net.ListenTCP("tcp", tcpServer)

	for {
		// new connection from client
		conn, err := listener.Accept()
		log.Println("new connection")
		if err != nil {
			log.Println(err)
			continue
		}
		s.currentConn = conn
		go s.readFromClient()
	}
}

func (s *TCPServer) GetBuffChan() chan []byte {
	return s.buffchan
}

func (s *TCPServer) Send(b []byte) (err error) {
	_, err = s.currentConn.Write(b)
	return
}

func (s *TCPServer) readFromClient() {
	log.Println("start reading from client")
	buffer := make([]byte, config.TransferBlockSize)
	for {
		n, err := s.currentConn.Read(buffer)
		if err != nil {
			// client send done
			common.ErrorHandleDebug(err)
			return
		}
		dataString := string(buffer[0:n])

		if len(dataString) != 0 {
			s.buffchan <- buffer[0:n]
		}
	}
}
