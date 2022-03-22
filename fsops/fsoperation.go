package fsops

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
)

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
	return
}

func Delete(path string) (err error) {
	return
}

func Create(path string) (err error) {
	return
}