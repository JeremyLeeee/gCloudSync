package common

import (
	"log"
	"os"
	"strconv"
)

type FileOp int

const (
	OpCreate FileOp = 1 << (32 - 1 - iota)
	OpRemove
	OpModify
	OpRename
)

const (
	AcceptOK = "AcceptOK"
	Done     = "Done"
)

type FsEvent struct {
	Op       FileOp
	FileName string
	IsDir    bool
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

func ErrorHandleFatal(err error) {
	if err != nil {
		log.Println("Fatal error: ", err)
		os.Exit(-1)
	}
}

func ErrorHandleDebug(err error) {
	if err != nil {
		log.Println("Error: ", err)
	}
}
