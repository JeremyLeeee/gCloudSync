package main

import (
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/core"
)

var logtag string = "[Main]"

func main() {
	common.PrintLogo()
	sc := core.NewServerCore(config.ServerRootPath)
	sc.StartServer()
}
