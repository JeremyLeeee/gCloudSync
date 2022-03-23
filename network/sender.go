package network

type ISender interface {
	Send(b []byte) error
	GetChan() chan []byte
}

type Sender struct {
	destAddr string
	port     string
	buffchan chan []byte
}

func NewSender(ip string, port string) ISender {
	buffchan := make(chan []byte)
	return &Sender{destAddr: ip, port: port, buffchan: buffchan}
}

func (s *Sender) Send(b []byte) error {
	return nil
}

func (s *Sender) GetChan() chan []byte {
	return nil
}
