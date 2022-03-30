package main

import (
	"gcloudsync/common"
	"gcloudsync/fswatcher"
	"gcloudsync/metadata"
	"gcloudsync/processer"
	"log"
	"os"
)

func main() {
	processer.StartClient()
}

func testHeaderTrans() {
	header := metadata.NewHeader(12, common.SysDone)
	buf, err := header.ToByteArray()
	common.ErrorHandleDebug(err)
	log.Println(buf)

	tag, datalen, err := metadata.GetHeaderFromData(buf)
	common.ErrorHandleDebug(err)

	log.Println("tag:", tag, "datalen:", datalen)
}

func testFsWatcher() {
	path := os.Args[1]
	fw := fswatcher.NewFsWatcher(path)
	fschan := fw.GetChan()

	// start watching
	go fw.StartWatching()

	log.Println("start watching folder:", path)

	// dispatch fs event
	for {
		event := <-fschan
		log.Print(event)
	}
}
