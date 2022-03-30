package dispatcher

import "gcloudsync/config"

type IDispatcher interface {
	SetData([]byte)
	GetData() []byte
	GetDispatchedData() []byte
}

// handle data format transfer, event dispatcher
// return data needed to be transfered
// apply fsops if needed
type Dispatcher struct {
	data []byte
}

func NewDispatcher() IDispatcher {
	data := make([]byte, config.TransferBlockSize)
	return &Dispatcher{data: data}
}

func (d *Dispatcher) SetData(b []byte) {
	d.data = b
}

func (d *Dispatcher) GetData() []byte {
	return d.data
}

func (d *Dispatcher) GetDispatchedData() []byte {
	return nil
}
