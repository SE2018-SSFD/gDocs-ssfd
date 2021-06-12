package dfs

import (
	"io/ioutil"
	"os"
	"sync"
)

const root = "../../../mockdfs"

var fdMap sync.Map
var curFd = 0

func mockOpen(path string) (int, error) {
	file, err := os.Open(root + path)
	fdMap.Store(curFd, file)
	curFd += 1
	return curFd - 1, err
}

func mockCreate(path string) (int, error) {
	file, err := os.Create(root + path)
	fdMap.Store(curFd, file)
	curFd += 1
	return curFd - 1, err
}

func mockDelete(path string) error {
	err := os.Remove(root + path)
	return err
}

func mockRead(fd int, off int64, len int64) (string, error) {
	f, _ := fdMap.Load(fd)
	file := f.(os.File)
	raw := make([]byte, len)
	n, err := file.ReadAt(raw, off)

	return string(raw[0:n]), err
}

func mockWrite(fd int, off int64, content string) (int64, error) {
	f, _ := fdMap.Load(fd)
	file := f.(os.File)
	n, err := file.WriteAt([]byte(content), off)

	return int64(n), err
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

func osFileInfo2DfsFileInfo(before os.FileInfo) FileInfo {
	return FileInfo{
		IsDir: before.IsDir(),
		Size: before.Size(),
	}
}