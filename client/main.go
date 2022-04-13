package main

import "gcloudsync/rsync"

var logtag string = "[Main]"

func main() {
	rsync.GenerateHashTable("/Users/jeremylee/Documents/code/handy/handy.zip")
	// log.Println(logtag, "Root Path:", config.ClientRootPath)
	// cc := core.NewClientCore(config.ClientRootPath)
	// cc.StartClient()
}
