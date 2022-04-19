package config

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"gcloudsync/internal/common"
	"log"
	"os"
	"sync"
)

var logtag string = "[Config]"

type Config struct {
	ServerIP          string
	TruncateBlockSize int
	TransferBlockSize int
	RootPath          string
}

type ServerRoot struct {
	RootPath string
}

// configurable
var ServerIP string = "127.0.0.1"
var TruncateBlockSize int = 1024
var TransferBlockSize int = 1024 * 4
var MaxBufferSize int = 1024 * 1024 * 192
var ClientRootPath string = "./"

// un-configurable
var Port string = "8909"
var BuffChanSize int = 1000
var EventChanSize int = 1000
var ServerRootPath string = "./"

var config *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		config = new(Config)
	})
	return config
}

func (c *Config) ReadConfigFromJson(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return errors.New("open json failed")
	}
	defer file.Close()

	fileinfo, err := os.Stat(path)
	common.ErrorHandleDebug(logtag, err)
	filesize := fileinfo.Size()

	data := make([]byte, filesize)
	_, err = file.Read(data)
	common.ErrorHandleDebug(logtag, err)

	err = json.Unmarshal(data, c)

	c.changeGlobalConfigStatus()

	PrintCurrentConfig()
	return err
}

func (c *Config) changeGlobalConfigStatus() {
	ServerIP = c.ServerIP
	TruncateBlockSize = c.TruncateBlockSize
	TransferBlockSize = c.TransferBlockSize
	ClientRootPath = c.RootPath
}

func (c *Config) ToBytes() []byte {
	b := make([]byte, 4)
	var buf bytes.Buffer

	binary.BigEndian.PutUint32(b, uint32(c.TruncateBlockSize))
	buf.Write([]byte(b))
	binary.BigEndian.PutUint32(b, uint32(c.TransferBlockSize))

	buf.Write([]byte(b))

	return buf.Bytes()
}

func (c *Config) ConfigFromBytes(b []byte) {
	c.TruncateBlockSize = int(binary.BigEndian.Uint32(b[0:4]))
	c.TransferBlockSize = int(binary.BigEndian.Uint32(b[4:8]))
	c.changeGlobalConfigStatus()
}

func PrintCurrentConfig() {
	log.Println(logtag, "Current Configuration:")
	if ServerIP != "" {
		log.Println(logtag, "ServerIP:", ServerIP)
	}
	log.Println(logtag, "TruncateBlockSize:", TruncateBlockSize)
	log.Println(logtag, "TransferBlockSize:", TransferBlockSize)
	log.Println(logtag, "MaxBufferSize:", MaxBufferSize)
	if ClientRootPath != "" {
		log.Println(logtag, "ClientRootPath:", ClientRootPath)
	}
}

func ConfigServerRootPath(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return errors.New("open json failed")
	}
	defer file.Close()

	fileinfo, err := os.Stat(path)
	common.ErrorHandleDebug(logtag, err)
	filesize := fileinfo.Size()

	data := make([]byte, filesize)
	_, err = file.Read(data)
	common.ErrorHandleDebug(logtag, err)

	s := ServerRoot{}
	err = json.Unmarshal(data, &s)
	ServerRootPath = s.RootPath
	return err
}
