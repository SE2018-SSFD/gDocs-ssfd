package dfs

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

const root = "/mockdfs"

var fdMap sync.Map
var curFd = 0

func mockOpen(path string) (fd int, err error) {
	file, err := os.OpenFile(root + path, os.O_RDWR, os.ModePerm)
	fdMap.Store(curFd, file)
	fd = curFd
	curFd += 1
	return fd, err
}

func mockClose(fd int) (err error) {
	f, _ := fdMap.Load(fd)
	file := f.(*os.File)

	if err = file.Close(); err != nil {
		return err
	}

	fdMap.Delete(fd)
	return nil
}

func mockCreate(path string) (int, error) {
	err := mockNameX(root + path, true)
	if err != nil {
		return 0, err
	}
	file, err := os.Create(root + path)
	fdMap.Store(curFd, file)
	curFd += 1
	return curFd - 1, err
}

func mockDelete(path string) error {
	err := os.RemoveAll(root + path)
	return err
}

func mockRead(fd int, off int64, len int64) (string, error) {
	f, _ := fdMap.Load(fd)
	file := f.(*os.File)
	raw := make([]byte, len)
	n, err := file.ReadAt(raw, off)

	return string(raw[0:n]), err
}

func mockWrite(fd int, off int64, content string) (int64, error) {
	f, _ := fdMap.Load(fd)
	file := f.(*os.File)
	n, err := file.WriteAt([]byte(content), off)

	return int64(n), err
}

func mockTruncate(fd int, length int64) error {
	f, _ := fdMap.Load(fd)
	file := f.(*os.File)
	return file.Truncate(length)
}

func mockScan(path string) ([]FileInfo, error) {
	osFileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dfsFileInfos []FileInfo
	for _, info := range osFileInfos {
		dfsFileInfos = append(dfsFileInfos, osFileInfo2DfsFileInfo(info))
	}

	return dfsFileInfos, nil
}

func mockStat(path string) (FileInfo, error) {
	stat, err := os.Stat(root + path)
	if err != nil {
		return FileInfo{}, err
	}

	fileInfo := osFileInfo2DfsFileInfo(stat)
	return fileInfo, nil
}

func mockNameX(path string, create bool) error {
	if create == true {
		dirPath := path[:strings.LastIndex(path, "/")]
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func osFileInfo2DfsFileInfo(before os.FileInfo) FileInfo {
	return FileInfo{
		Name: before.Name(),
		IsDir: before.IsDir(),
		Size: before.Size(),
	}
}