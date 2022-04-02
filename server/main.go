package main

import (
	"gcloudsync/config"
	"gcloudsync/core"
	"log"
)

var logtag string = "[Main]"

func main() {
	log.Println(logtag, "Root Path:", config.ServerRootPath)
	sc := core.NewServerCore(config.ServerRootPath)
	sc.StartServer()
}
