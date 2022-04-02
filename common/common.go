package common

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"strconv"
)

type FsOp int
type SysOp uint16

// fs event
const (
	OpCreate FsOp = 1 << (32 - 1 - iota)
	OpRemove
	OpModify
	OpRename
)

const (
	AcceptOK = "AcceptOK"
	Done     = "Done"
)

const (
	IsLastPackage   = 51
	IsNotLastPacage = 52
)

// system event
const (
	SysConnected SysOp = 101 + iota
	SysDone
	SysCheckConsistence
	SysInit
)

type FsEvent struct {
	Op         FsOp
	FileName   string
	IsDir      bool
	OriginFile string // for rename event
}

func (fe FsEvent) String() string {
	var eventString string

	switch fe.Op {
	case OpCreate:
		eventString = "create"
	case OpModify:
		eventString = "modify"
	case OpRename:
		eventString = "rename"
	case OpRemove:
		eventString = "remove"
	}

	return eventString + ": " + fe.FileName + ", isdir: " + strconv.FormatBool(fe.IsDir)
}

func ErrorHandleFatal(tag string, err error) {
	if err != nil {
		log.Println(tag, "Fatal error: ", err)
		os.Exit(-1)
	}
}

func ErrorHandleDebug(tag string, err error) {
	if err != nil {
		log.Println(tag, "Error: ", err)
	}
}

func SysOpToByteArray(op SysOp) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, op); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MergeArray(b1 []byte, b2 []byte) []byte {
	var buf bytes.Buffer
	buf.Write(b1)
	buf.Write(b2)
	return buf.Bytes()
}
