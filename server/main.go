package main

import (
	"gcloudsync/config"
	"gcloudsync/core"
)

func main() {
	path := config.ServerRootPath
	sc := core.NewServerCore(path)
	sc.StartServer()
}
