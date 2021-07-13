package dfs

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

const root = "/mockdfs"

var fdMap sync.Map
var curFd = int32(0)

var NoFdErr = errors.New("mockDfs: fd does not exist")

func mockOpen(path string, isAppend bool) (fd int, err error) {
	flag := os.O_RDWR
	if isAppend {
		flag |= os.O_APPEND
	}
	file, err := os.OpenFile(root+path, flag, os.ModePerm)
	if err != nil {
		return 0, withStackedMessagef(err, "mockOpen fails")
	}

	fd = int(atomic.AddInt32(&curFd, 1))
	fdMap.Store(fd, file)

	return fd, nil
}

func mockClose(fd int) (err error) {
	f, _ := fdMap.Load(fd)
	file := f.(*os.File)

	if err = file.Close(); err != nil {
		return withStackedMessagef(err, "mockClose fails")
	}

	fdMap.Delete(fd)
	return nil
}

func mockCreate(path string) (int, error) {
	dirPath := (root + path)[:strings.LastIndex(root+path, "/")]
	err := mockNameX(dirPath, true)
	if err != nil {
		return 0, err
	}

	file, err := os.Create(root + path)
	if err != nil {
		return 0, withStackedMessagef(err, "mockCreate fails when creating")
	}

	fd := int(atomic.AddInt32(&curFd, 1))
	fdMap.Store(fd, file)

	return fd, nil
}

func mockMkdir(path string) (err error) {
	err = mockNameX(root+path, true)
	if err != nil {
		return err
	}

	return nil
}

func mockDelete(path string) error {
	err := os.RemoveAll(root + path)
	if err != nil {
		return withStackedMessagef(err, "mockDelete fails")
	}

	return nil
}

func mockRead(fd int, off int64, len int64, _ ...bool) ([]byte, error) {
	f, ok := fdMap.Load(fd)
	if !ok {
		return nil, withStackedMessagef(NoFdErr, "mockRead fails")
	}

	file := f.(*os.File)
	raw := make([]byte, len)
	n, err := file.ReadAt(raw, off)
	if err != nil {
		return nil, withStackedMessagef(err, "mockRead fails")
	}

	return raw[0:n], nil
}

func mockReadAll(path string, _ ...bool) (data []byte, err error) {
	file, err := os.OpenFile(root+path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, withStackedMessagef(err, "mockOpen fails")
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, withStackedMessagef(err, "mockReadAll fails when doing Stat")
	}

	raw := make([]byte, fileInfo.Size())
	n, err := file.Read(raw)
	if err != nil {
		return nil, withStackedMessagef(err, "mockReadAll fails when Reading")
	}

	return raw[0:n], nil
}

func mockWrite(fd int, off int64, data []byte) (int64, error) {
	f, ok := fdMap.Load(fd)
	if !ok {
		return 0, withStackedMessagef(NoFdErr, "mockWrite fails")
	}

	file := f.(*os.File)
	n, err := file.WriteAt(data, off)
	if err != nil {
		return 0, withStackedMessagef(err, "mockWrite fails")
	}

	return int64(n), nil
}

func mockAppend(fd int, data []byte) (bytesWritten int64, err error) {
	f, ok := fdMap.Load(fd)
	if !ok {
		return 0, NoFdErr
	}

	file := f.(*os.File)
	n, err := file.Write(data)
	if err != nil {
		return 0, withStackedMessagef(err, "mockAppend fails")
	}

	return int64(n), nil
}

func mockScan(path string) ([]FileInfo, error) {
	osFileInfos, err := ioutil.ReadDir(root + path)
	if err != nil {
		return nil, withStackedMessagef(err, "mockScan fails")
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
		return FileInfo{}, withStackedMessagef(err, "mockStat fails")
	}

	fileInfo := osFileInfo2DfsFileInfo(stat)
	return fileInfo, nil
}

func mockNameX(path string, create bool) error {
	if create == true {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return withStackedMessagef(err, "mockNameX fails")
		}
	}

	return nil
}

func osFileInfo2DfsFileInfo(before os.FileInfo) FileInfo {
	return FileInfo{
		Name:  before.Name(),
		IsDir: before.IsDir(),
	}
}
