package common

type FileOp int

const (
	OpCreate FileOp = 1 << (32 - 1 - iota)
	OpRemove
	OpWrite
	OpRename
)
