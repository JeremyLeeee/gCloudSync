package fsops

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
)

var currentFile *os.File
var isOpened bool = false

func IsFileExist(path string) bool {
	file, err := os.Open(path)
	result := true
	if err != nil {
		// log.Println(filename + " not exist")
		result = false
	}
	defer file.Close()
	return result
}

func IsFolder(path string) (bool, error) {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileinfo.IsDir(), err
}

func GetFileList(path string) (flist []string, err error) {
	if ok, _ := IsFolder(path); !ok {
		return flist, errors.New("not a folder")
	}

	filist, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("read dir error")
	}

	for _, f := range filist {
		fname := path + "/" + f.Name()
		flist = append(flist, fname)
	}

	return
}

func GetSubDirs(path string) (dirlist []string, err error) {
	if ok, _ := IsFolder(path); !ok {
		return dirlist, errors.New("not a folder")
	}

	filist, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("read dir error")
	}

	for _, f := range filist {
		if f.IsDir() {
			fname := path + "/" + f.Name()
			dirlist = append(dirlist, fname)
		}
	}

	return
}

func Makedir(path string) (err error) {
	return os.Mkdir(path, 0666)
}

func Delete(path string) (err error) {
	return os.RemoveAll(path)
}

func Create(path string) (err error) {
	if !IsFileExist(path) {
		_, err := os.Create(path)
		return err
	}
	return nil
}

func WriteOnce(path string, b []byte, off int64) (n int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return -1, err
	}
	defer file.Close()
	return file.WriteAt(b, off)
}

func ReadOnce(path string, b []byte, off int64) (n int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return -1, err
	}
	defer file.Close()
	return file.ReadAt(b, off)
}

func Write(path string, b []byte, off int64) (n int, err error) {
	if !isOpened {
		currentFile, err = os.Open(path)
		if err != nil {
			return -1, err
		}
		isOpened = true
	}
	return currentFile.WriteAt(b, off)
}

func Read(path string, b []byte, off int64) (n int, err error) {
	if !isOpened {
		currentFile, err = os.Open(path)
		if err != nil {
			return -1, err
		}
		isOpened = true
	}
	return currentFile.ReadAt(b, off)
}

// do not forget to call it after read or write
func CloseCurrentFile() {
	currentFile.Close()
	isOpened = false
}
