package common

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

var logtag string = "[Common]"

var Version string = "0.1.0"

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
	OpChmod
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
	SysInitUpload
	SysInitSyncConfig
	SysInitSyncFolder
	SysInitSyncFile
	SysInitFinished
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
	SysOpChmod
)

type FsEvent struct {
	Op         FsOp
	FileName   string
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
	case OpChmod:
		eventString = "chmod"
	}

	return eventString + ": " + fe.FileName
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

func PrintLogo() {
	fmt.Println("             ________                _______                 ")
	fmt.Println("      ____ _/ ____/ /___  __  ______/ / ___/__  ______  _____")
	fmt.Println("     / __ `/ /   / / __ \\/ / / / __  /\\__ \\/ / / / __ \\/ ___/")
	fmt.Println("    / /_/ / /___/ / /_/ / /_/ / /_/ /___/ / /_/ / / / / /__  ")
	fmt.Println("    \\__, /\\____/_/\\____/\\__,_/\\__,_//____/\\__, /_/ /_/\\___/  ")
	fmt.Println("   /____/                                /____/         ")
	fmt.Println("                                                version:", Version)
}
