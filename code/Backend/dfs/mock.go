package dfs

import (
	"os"
	"sync"
)

const root = "/Users/wukanzhen/SJTU/Lessons/云计算系统设计与开发/lab5/mockdfs"

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

func mockList(path string) ([]string, error) {
	return nil, nil
}

func mockStat(path string) (FileInfo, error) {
	stat, err := os.Stat(root + path)
	if err != nil {
		return FileInfo{}, err
	}

	fileInfo := FileInfo{
		IsDir: stat.IsDir(),
		Size: stat.Size(),
	}
	return fileInfo, nil
}