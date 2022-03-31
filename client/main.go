package main

import (
	"gcloudsync/core"
	"os"
)

func main() {
	cc := core.NewClientCore(os.Args[1])
	cc.StartClient()
}
