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
	buffchan := make(chan []byte, config.BuffChanSize)
	return &TCPServer{port: port, buffchan: buffchan}
}

// when new connection arrive, create an goroutine to handle receiving data.
// data from client will be writen to buffchan
func (s *TCPServer) Listen() {
	tcpServer, _ := net.ResolveTCPAddr("tcp4", ":"+config.Port)
	listener, _ := net.ListenTCP("tcp", tcpServer)

	for {
		// new connection from client
		conn, err := listener.Accept()
		log.Println(logtag, "new connection from:", conn.RemoteAddr())
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
	if len(b) <= config.TransferBlockSize {
		_, err = s.currentConn.Write(b)
		common.ErrorHandleDebug(logtag, err)
	} else {
		i := 0
		for ; i+config.TransferBlockSize <= len(b); i = i + config.TransferBlockSize {
			_, err = s.currentConn.Write(b[i : i+config.TransferBlockSize])
			common.ErrorHandleDebug(logtag, err)
		}
		_, err = s.currentConn.Write(b[i:])
	}

	return
}

func (s *TCPServer) readFromClient() {
	for {
		buffer := make([]byte, config.MaxBufferSize)
		n, err := s.currentConn.Read(buffer)
		if err != nil {
			// client send done
			common.ErrorHandleDebug(logtag, err)
			return
		}

		buff := buffer[0:n]
		if len(buff) != 0 {
			// log.Println(logtag, "receive len:", len(buff))
			s.buffchan <- buff
		}
	}
}
