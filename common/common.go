package common

import "strconv"

type FileOp int

const (
	OpCreate FileOp = 1 << (32 - 1 - iota)
	OpRemove
	OpModify
	OpRename
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
