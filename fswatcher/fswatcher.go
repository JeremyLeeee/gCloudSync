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
	fileMap map[string]int
	fschan  chan common.FsEvent
	watcher *fsnotify.Watcher
}

func NewFsWatcher(path string) *FsWatcher {
	f := new(FsWatcher)
	// set monitor path
	err := f.setPath(path)
	common.ErrorHandleFatal(logtag, err)
	// add all file to map
	f.fileMap = make(map[string]int)
	flist := fsops.GetAllFile(path)

	for _, file := range flist {
		f.fileMap[file] = 1
	}

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

func (f *FsWatcher) getEvent(fsevent common.FsEvent) (common.FsEvent, error) {
	flist := fsops.GetAllFile(f.path)

	if len(f.fileMap) == len(flist) {
		// rename or modify
		if fsevent.Op == common.OpModify {
			return fsevent, nil
		} else if fsevent.Op == common.OpRename {
			// rename
			// find new file name
			for _, file := range flist {
				if f.fileMap[file] == 0 {
					fsevent.OriginFile = fsevent.FileName
					fsevent.FileName = file
					return fsevent, nil
				}
			}
		} else {
			return fsevent, errors.New("invalid event")
		}
	} else if len(f.fileMap) < len(flist) {
		// create
		return fsevent, nil
	} else if len(f.fileMap) > len(flist) {
		// delete
		fsevent.Op = common.OpRemove
		return fsevent, nil
	}

	return fsevent, errors.New("invalid event")
}

func (f *FsWatcher) updateFileMap() {
	f.fileMap = make(map[string]int)
	flist := fsops.GetAllFile(f.path)

	for _, file := range flist {
		f.fileMap[file] = 1
	}
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
				fsevent, err := f.toFsEvent(event)
				common.ErrorHandleDebug(logtag, err)

				if fseventNew, err := f.getEvent(fsevent); err == nil {
					f.fschan <- fseventNew
					f.updateFileMap()
				}

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
