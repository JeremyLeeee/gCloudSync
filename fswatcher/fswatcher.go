package fswatcher

import (
	"errors"
	"gcloudsync/common"
	"gcloudsync/fsops"

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
	} else if event.Op&fsnotify.Rename == fsnotify.Rename {
		op = common.OpRename
	} else if event.Op&fsnotify.Write == fsnotify.Write {
		op = common.OpModify
	} else if event.Op&fsnotify.Remove == fsnotify.Remove {
		op = common.OpRemove
	} else if event.Op&fsnotify.Chmod == fsnotify.Chmod {
		op = common.OpChmod
	} else {
		return common.FsEvent{}, errors.New("unknown event")
	}

	return common.FsEvent{Op: op, FileName: event.Name}, nil
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
		var lastEvent common.FsEvent
		for {
			select {
			case event, ok := <-f.watcher.Events:
				if !ok {
					return
				}
				fsevent, err := f.toFsEvent(event)
				common.ErrorHandleDebug(logtag, err)

				// due to the bug of fsnotify
				// some events should be detected as combination below
				// rename: 1.create, 2.rename
				// remove: 1.chmod, 2.rename

				if lastEvent.Op == common.OpCreate {
					if fsevent.Op == common.OpRename {
						// if it is rename event
						// merge two events into one
						fsevent.OriginFile = fsevent.FileName
						fsevent.FileName = lastEvent.FileName
						f.fschan <- fsevent
					} else {
						f.fschan <- lastEvent
						f.fschan <- fsevent
					}
				} else {
					if fsevent.Op != common.OpCreate && fsevent.Op != common.OpRename {
						f.fschan <- fsevent
					}
				}

				if lastEvent.Op == common.OpChmod && fsevent.Op == common.OpRename {
					fsevent.Op = common.OpRemove
					f.fschan <- fsevent
				}

				lastEvent = fsevent

			case err, ok := <-f.watcher.Errors:
				if !ok {
					return
				}
				common.ErrorHandleDebug(logtag, err)
			}
		}
	}()

	err := f.addAll()
	common.ErrorHandleFatal(logtag, err)
	<-done
}
