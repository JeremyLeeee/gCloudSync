package main

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/core"
)

var logtag string = "[Main]"

func main() {
	common.PrintLogo()
	err := config.ConfigServerRootPath("./config.json")
	common.ErrorHandleFatal(logtag, err)
	sc := core.NewServerCore(config.ServerRootPath)
	sc.StartServer()
}
