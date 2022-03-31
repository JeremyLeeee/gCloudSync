package fswatcher

import (
	"errors"
	"gcloudsync/common"
	"gcloudsync/fsops"
	"log"

	"github.com/fsnotify/fsnotify"
)

var logtag string = "[FsWatcher]"

type FsWatcher struct {
	path    string
	fschan  chan common.FsEvent
	watcher *fsnotify.Watcher
}

func NewFsWatcher(path string) *FsWatcher {
	f := new(FsWatcher)
	// set monitor path
	err := f.setPath(path)
	common.ErrorHandleFatal(logtag, err)
	// get a watcher
	f.watcher, err = fsnotify.NewWatcher()
	common.ErrorHandleDebug(logtag, err)
	// make a channal for communication
	ch := make(chan common.FsEvent)
	f.fschan = ch
	return f
}

func (f *FsWatcher) setPath(path string) (err error) {
	// check path availibitiy
	ok := fsops.IsFileExist(path)
	if !ok {
		return errors.New("file not exist")
	}
	f.path = path
	return nil
}

func (f *FsWatcher) GetChan() chan common.FsEvent {
	return f.fschan
}

// since fsnotify does not work well on delete action on macOS
// we transfer fsnotify event to our own event
func (f *FsWatcher) toFsEvent(event fsnotify.Event) (common.FsEvent, error) {
	var op common.FsOp
	var isdir bool
	if ok, _ := fsops.IsFolder(event.Name); ok {
		isdir = true
	} else {
		isdir = false
	}

	if event.Op&fsnotify.Create == fsnotify.Create {
		op = common.OpCreate
		// if it is dir, add to watcher
		if isdir {
			f.addDir(event.Name)
		}
	} else if event.Op&fsnotify.Remove == fsnotify.Remove {
		op = common.OpRemove
	} else if event.Op&fsnotify.Write == fsnotify.Write {
		op = common.OpModify
	} else if event.Op&fsnotify.Rename == fsnotify.Rename {
		ok := fsops.IsFileExist(event.Name)
		if !ok {
			op = common.OpRemove
		} else {
			op = common.OpRename
		}
	} else {
		return common.FsEvent{}, errors.New("unknown event")
	}

	return common.FsEvent{Op: op, FileName: event.Name, IsDir: isdir}, nil
}

func (f *FsWatcher) addAll() (err error) {
	f.addDir(f.path)
	return nil
}

// add all subdir recursively
func (f *FsWatcher) addDir(path string) (err error) {
	if ok, _ := fsops.IsFolder(path); !ok {
		return errors.New("not a folder")
	}
	f.watcher.Add(path)

	flist, err := fsops.GetSubDirs(path)
	common.ErrorHandleDebug(logtag, err)

	for _, folder := range flist {
		f.addDir(folder)
	}
	return
}

func (f *FsWatcher) StartWatching() {
	defer f.watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-f.watcher.Events:
				if !ok {
					return
				}
				// log.Println("event:", event)
				fsevent, err := f.toFsEvent(event)
				if err == nil {
					f.fschan <- fsevent
				}
			case err, ok := <-f.watcher.Errors:
				if !ok {
					return
				}
				log.Println(logtag, "err:", err)
			}
		}
	}()

	err := f.addAll()
	common.ErrorHandleFatal(logtag, err)
	<-done
}
