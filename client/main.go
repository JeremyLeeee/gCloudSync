package main

import (
	"gcloudsync/config"
	"gcloudsync/core"
)

func main() {
	cc := core.NewClientCore(config.ClientRootPath)
	cc.StartClient()
}
