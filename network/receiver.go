package network

type IReceiver interface {
	Listen() error
	Send(b []byte) error
	GetChan() chan []byte
}

type Receiver struct {
	ip       string
	port     string
	buffchan chan []byte
}

func NewReceiver(ip string, port string) IReceiver {
	buffchan := make(chan []byte)
	return &Receiver{ip: ip, port: port, buffchan: buffchan}
}

func (r *Receiver) Listen() error {
	return nil
}

func (r *Receiver) Send(b []byte) error {
	return nil
}

func (r *Receiver) GetChan() chan []byte {
	return nil
}
