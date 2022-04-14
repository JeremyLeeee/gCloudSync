package common

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"log"
	"os"
	"strconv"
)

var logtag string = "[Common]"

type FsOp int
type SysOp uint16

// fs event
const (
	OpCreate FsOp = 200 + iota
	OpRemove
	OpModify
	OpRename
	OpFetch // need sync file from server
	OpMkdir
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
	SysSyncFolder
	SysSyncFileBegin
	SysSyncFileEmpty
	SysSyncFileNotEmpty
	SysSyncFileDirect
	SysSyncFinished
	SysSyncGenerateDiff
	SysSyncReformFile

	SysOpRemove
	SysOpCreate
	SysOpModify
	SysOpRename
	SysOpMkdir
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
	case OpFetch:
		eventString = "fetch"
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

func GetFileMd5(path string) []byte {
	file, err := os.Open(path)
	ErrorHandleFatal(logtag, err)
	md5h := md5.New()
	io.Copy(md5h, file)
	result := md5h.Sum([]byte{})
	return result
}

func GetByteMd5(b []byte) []byte {
	md5h := md5.New()
	md5h.Write(b)
	result := md5h.Sum([]byte{})
	return result
}

func TODO(str string) {
	log.Println(logtag, "TODO:", str)
}
