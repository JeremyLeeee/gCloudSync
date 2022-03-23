package main

import (
	"gcloudsync/fswatcher"
	"log"
	"os"
)

func main() {
	path := os.Args[1]
	fw := fswatcher.NewFsWatcher(path)
	fschan := fw.GetChan()

	// start watching
	go func() {
		fw.StartWatching()
	}()
	log.Println("start watching folder:", path)

	// dispatch fs event
	for {
		event := <-fschan
		log.Print(event)
	}
}
