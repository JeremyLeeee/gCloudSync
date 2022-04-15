package fsops

import (
	"errors"
	"gcloudsync/common"
	"gcloudsync/config"
	"io/ioutil"
	"log"
	"os"
)

var currentFile *os.File
var isOpened bool = false
var logtag string = "[FsOps]"

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
	common.ErrorHandleDebug(logtag, err)

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
		log.Println(logtag, "read dir error")
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
	if !IsFileExist(path) {
		log.Println(logtag, "mkdir:", path)
		return os.Mkdir(path, 0777)
	}
	return nil
}

func Delete(path string) (err error) {
	return os.RemoveAll(path)
}

func Create(path string) (err error) {
	if !IsFileExist(path) {
		file, err := os.Create(path)
		file.Close()
		return err
	}
	return nil
}

func Rename(old string, new string) (err error) {
	return os.Rename(old, new)
}

func WriteOnce(path string, b []byte, off int64) (n int, err error) {
	file, err := os.OpenFile(path, os.O_WRONLY, 0777)
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
	if !IsFileExist(path) {
		Create(path)
	}
	if !isOpened {
		currentFile, err = os.OpenFile(path, os.O_WRONLY, 0777)
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

func ReadAll(path string) (b []byte, err error) {
	filesize, err := GetFileSize(path)
	common.ErrorHandleDebug(logtag, err)
	databuff := make([]byte, filesize)

	offset := 0
	for {
		n, err := Read(path, databuff[offset:], int64(offset))
		common.ErrorHandleDebug(logtag, err)
		if n > 0 {
			offset = offset + n
		} else {
			break
		}
	}
	CloseCurrentFile()
	return databuff, err
}

func WriteAll(path string, b []byte) (err error) {
	offset := 0
	for {
		n, err := Write(path, b[offset:], int64(offset))
		common.ErrorHandleDebug(logtag, err)
		if n < len(b)-offset {
			offset = offset + n
		} else {
			break
		}
	}
	CloseCurrentFile()
	return err
}
func WriteAllAt(path string, b []byte, offset int) (err error) {
	count := 0
	for {
		n, err := Write(path, b[count:], int64(offset))
		common.ErrorHandleDebug(logtag, err)
		if n < len(b)-offset {
			offset = offset + n
			count = count + n
		} else {
			break
		}
	}
	CloseCurrentFile()
	return err
}

// do not forget to call it after read or write
func CloseCurrentFile() {
	currentFile.Close()
	isOpened = false
}

func GetAllFile(path string) (result []string) {
	return getAllFileHelper(path, result)
}

func getAllFileHelper(path string, result []string) []string {
	result = append(result, path)
	if ok, _ := IsFolder(path); !ok {
		return result
	}

	filist, err := ioutil.ReadDir(path)
	common.ErrorHandleDebug(logtag, err)

	for _, file := range filist {
		result = getAllFileHelper(path+"/"+file.Name(), result)
	}
	return result
}

func RemoveRootPrefix(path string, isClient bool) (result string) {
	// assume all input include root path
	b := []byte(path)
	if isClient {
		b = b[len(config.ClientRootPath):]
	} else {
		b = b[len(config.ServerRootPath):]
	}
	return string(b)
}

func GetFileSize(path string) (size int64, err error) {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileinfo.Size(), nil
}
