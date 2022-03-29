package network

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"net"
)

type ITCPClient interface {
	Connect() error
	Send(b []byte) error
	Receive() []byte
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
	buffchan := make(chan []byte)
	return &TCPClient{destAddr: ip, port: port, buffchan: buffchan}
}

func (c *TCPClient) Connect() error {
	// connect to server
	addr, err := net.ResolveTCPAddr("tcp4", config.ServerIP+":"+config.Port)
	common.ErrorHandleFatal(err)

	c.conn, err = net.DialTCP("tcp4", nil, addr)
	common.ErrorHandleFatal(err)

	return err
}

func (c *TCPClient) Send(b []byte) (err error) {
	_, err = c.conn.Write(b)
	common.ErrorHandleDebug(err)
	return err
}

func (c *TCPClient) GetBuffChan() chan []byte {
	return c.buffchan
}

func (c *TCPClient) Receive() []byte {
	buffer := make([]byte, config.TransferBlockSize)
	c.conn.Read(buffer)
	return buffer
}

func (c *TCPClient) Close() {
	c.conn.Close()
}
