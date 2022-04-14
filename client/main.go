package main

import (
	"gcloudsync/config"
	"gcloudsync/core"
	"log"
)

var logtag string = "[Main]"

func main() {
	log.Println(logtag, "Root Path:", config.ClientRootPath)
	cc := core.NewClientCore(config.ClientRootPath)
	cc.StartClient()
}
