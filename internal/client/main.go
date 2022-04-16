package main

import (
	"gcloudsync/internal/common"
	"gcloudsync/internal/config"
	"gcloudsync/internal/core"
	"log"
)

var logtag string = "[Main]"

func main() {
	common.PrintLogo()
	// get config from json
	cg := config.GetConfig()
	err := cg.ReadConfigFromJson("./config.json")
	if err != nil {
		err := cg.ReadConfigFromJson("../config.json")
		if err != nil {
			log.Panicln(logtag, "unable to process config.json.")
		}
	}
	// start client
	cc := core.NewClientCore(config.ClientRootPath)
	cc.StartClient()
}
