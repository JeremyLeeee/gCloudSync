package fswatcher

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FsWatcher struct {
	path   string
	fschan chan fsnotify.Event
}

func NewFsWatcher() *FsWatcher {
	c := make(chan fsnotify.Event)
	return &FsWatcher{fschan: c}
}

func (f *FsWatcher) SetPath(path string) (err error) {
	// check path availibitiy
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("invalid path")
		file.Close()
		return err
	}
	file.Close()
	f.path = path
	return nil
}

func (f *FsWatcher) GetChan() chan fsnotify.Event {
	return f.fschan
}

func (f *FsWatcher) StartWatching() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Println("event:", event)
				f.fschan <- event
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("err:", err)
			}
		}
	}()

	err = watcher.Add(f.path)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
