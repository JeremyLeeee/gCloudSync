package network

import (
	"gcloudsync/internal/common"
	"gcloudsync/internal/config"
	"net"
)

var logtag string = "[Network]"

type ITCPClient interface {
	Connect() error
	Send(b []byte) error
	Receive() []byte
	ReadFromServer()
	GetBuffChan() chan []byte
	Close()
}

type TCPClient struct {
	destAddr string
	port     string
	buffchan chan []byte
	conn     *net.TCPConn
}

func NewClient(ip string, port string) ITCPClient {
	buffchan := make(chan []byte, config.BuffChanSize)
	return &TCPClient{destAddr: ip, port: port, buffchan: buffchan}
}

func (c *TCPClient) Connect() error {
	// connect to server
	addr, err := net.ResolveTCPAddr("tcp4", config.ServerIP+":"+config.Port)
	common.ErrorHandleFatal(logtag, err)

	c.conn, err = net.DialTCP("tcp4", nil, addr)
	common.ErrorHandleFatal(logtag, err)

	return err
}

func (c *TCPClient) Send(b []byte) (err error) {
	if len(b) <= config.TransferBlockSize {
		_, err = c.conn.Write(b)
		common.ErrorHandleDebug(logtag, err)
	} else {
		i := 0
		for ; i+config.TransferBlockSize <= len(b); i = i + config.TransferBlockSize {
			_, err = c.conn.Write(b[i : i+config.TransferBlockSize])
			common.ErrorHandleDebug(logtag, err)
		}
		_, err = c.conn.Write(b[i:])
	}
	return
}

func (c *TCPClient) GetBuffChan() chan []byte {
	return c.buffchan
}

func (c *TCPClient) Receive() []byte {
	buffer := make([]byte, config.MaxBufferSize)
	c.conn.Read(buffer)
	return buffer
}

func (c *TCPClient) Close() {
	c.conn.Close()
}

func (c *TCPClient) ReadFromServer() {
	for {
		buffer := make([]byte, config.MaxBufferSize)
		n, err := c.conn.Read(buffer)
		if err != nil {
			// client send done
			common.ErrorHandleDebug(logtag, err)
			return
		}

		buff := buffer[0:n]
		if len(buff) != 0 {
			// log.Println(logtag, "receive: "+string(buff))
			c.buffchan <- buff
		}
	}
}
