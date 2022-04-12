package config

const (
	ServerIP          = "127.0.0.1"
	Port              = "8888"
	TruncateBlockSize = 1024
	TransferBlockSize = 1024 * 4
	MaxBufferSize     = 1024 * 1024 * 16
	BuffChanSize      = 100
	EventChanSize     = 1000
)

var ServerRootPath string = "/Users/jeremylee/Documents/code/gcs_root/server"
var ClientRootPath string = "/Users/jeremylee/Documents/code/gcs_root/client"
