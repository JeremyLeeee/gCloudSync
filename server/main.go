package main

import (
	"gcloudsync/core"
)

func main() {
	sc := core.NewServerCore()
	sc.StartServer()
}
