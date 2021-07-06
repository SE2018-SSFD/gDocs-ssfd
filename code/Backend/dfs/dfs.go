package dfs

import (
	"backend/utils/logger"
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	InvalidPathErr		=	errors.New("Path is Invalid")
	InvalidFdErr		=	errors.New("FD is Invalid")

	OpenErr				=	errors.New("Cannot open file")
	StatNonExistentErr	=	errors.New("File is nonexistent")
)


// interfaces
func dfsOpen(path string) (fd int, err error) {
	absPath := checkPath(path)
	if absPath == "" {
		return 0, withStackedMessagef(InvalidPathErr, "[%s] dfsOpen", path)
	}

	reqBody := OpenArg{
		Path: absPath,
	}
	respBody := OpenRet{}

	err = post("open", reqBody, &respBody)
	if err != nil {
		return 0, withStackedMessagef(err, "[%s] dfsOpen", path)
	}

	if respBody.Fd == 0 {
		return 0, withStackedMessagef(OpenErr, "[%s] dfsOpen", path)
	}

	return respBody.Fd, err
}

func dfsClose(fd int) (err error) {
	if fd <= 0 {
		return withStackedMessagef(InvalidFdErr, "[%d] dfsClose", fd)
	}

	reqBody := CloseArg{
		Fd: fd,
	}

	err = post("close", reqBody, nil)
	if err != nil {
		return withStackedMessagef(err, "[%d] dfsClose", fd)
	}

	return nil
}

func dfsCreate(path string) (fd int, err error) {
	if path[0] != '/' {
		return 0, withStackedMessagef(InvalidPathErr, "[%s] dfsCreate", path)
	}

	// create all parents (NameX)
	levels := strings.Split(path[1:], "/")
	curPath := ""
	for _, l := range levels[:len(levels)-1] {
		curPath += "/" + l
		statReqBody := GetFileInfoArg{
			Path: transPath(curPath),
		}
		statRespBody := GetFileInfoRet{}

		err = post("fileInfo", statReqBody, &statRespBody)
		if err != nil {
			return 0, withStackedMessagef(err, "[%s] dfsCreate cannot stat parents \"%s\"", path, curPath)
		}

		if statRespBody.Exist {
			if !statRespBody.IsDir {
				return 0, withStackedMessagef(err, "[%s] dfsCreate parent \"%s\" is file", path, curPath)
			} else {
				continue
			}
		} else {
			mkdirReqBody := MkdirArg{
				Path: transPath(curPath),
			}

			err = post("mkdir", mkdirReqBody, nil)
			if err != nil {
				return 0, withStackedMessagef(err, "[%s] dfsCreate cannot create parents", path)
			}
		}
	}

	// create file
	createReqBody := CreateArg{
		Path: transPath(path),
	}

	err = post("create", createReqBody, nil)
	if err != nil {
		return 0, withStackedMessagef(err, "[%s] dfsCreate cannot create leaf file", path)
	}

	fd, err = dfsOpen(path)
	if err != nil {
		return 0, withStackedMessagef(err, "[%s] dfsCreate cannot open leaf file", path)
	}

	return fd, nil
}

// dfsDelete recursively delete all children if isDir
func dfsDelete(path string) (err error) {
	if path[0] != '/' {
		return withStackedMessagef(InvalidPathErr, "[%s] dfsDelete", path)
	}

	statReqBody := GetFileInfoArg{
		Path: transPath(path),
	}
	statRespBody := GetFileInfoRet{}

	err = post("fileInfo", statReqBody, &statRespBody)
	if err != nil {
		return withStackedMessagef(err, "[%s] dfsDelete cannot stat root", path)
	}

	if !statRespBody.IsDir {	// is file, simply delete
		deleteReqBody := DeleteArg{
			Path: transPath(path),
		}

		err = post("delete", deleteReqBody, nil)
		if err != nil {
			return withStackedMessagef(err, "[%s] dfsDelete cannot delete leaf file", path)
		}
	} else {					// is dir, recursively delete children, then delete self
		listReqBody := ListArg{
			Path: transPath(path),
		}
		listRespBody := ListRet{}

		err = post("list", listReqBody, &listRespBody)
		if err != nil {
			return withStackedMessagef(err, "[%s] dfsDelete cannot list root", path)
		}

		wg := sync.WaitGroup{}
		wg.Add(len(listRespBody.Files))
		for _, child := range listRespBody.Files {
			go func(child string) {
				_ = dfsDelete(path + "/" + child)
				wg.Done()
			}(child)
		}

		wg.Wait()

		deleteReqBody := DeleteArg{
			Path: transPath(path),
		}

		err = post("delete", deleteReqBody, nil)
		if err != nil {
			return withStackedMessagef(err, "[%s] dfsDelete cannot delete root self", path)
		}
	}

	return nil
}

func dfsRead(fd int, off int64, len int64) (content string, err error) {
	if fd <= 0 {
		return "", withStackedMessagef(InvalidFdErr, "[%d] dfsRead", strconv.Itoa(fd))
	}

	reqBody := ReadArg{
		Fd: fd,
		Offset: int(off),
		Len: int(len),
	}
	respBody := ReadRet{}

	err = post("read", reqBody, &respBody)
	if err != nil {
		return "", withStackedMessagef(err, "[%d] dfsRead", fd)
	}

	content = string(respBody.Data)
	return content, nil
}

func dfsWrite(fd int, off int64, content string) (bytesWritten int64, err error) {
	if fd <= 0 {
		return 0, withStackedMessagef(InvalidFdErr, "[%d] dfsWrite", strconv.Itoa(fd))
	}

	reqBody := WriteArg{
		Fd: fd,
		Offset: int(off),
		Data: []byte(content),
	}
	respBody := WriteRet{}

	err = post("write", reqBody, &respBody)
	if err != nil {
		return 0, withStackedMessagef(err, "[%d] dfsWrite", fd)
	}

	bytesWritten = int64(respBody.BytesWritten)
	return bytesWritten, nil
}

func dfsAppend(fd int, content string) (bytesWritten int64, err error) {
	if fd <= 0 {
		return 0, withStackedMessagef(InvalidFdErr, "[%d] dfsSppend", strconv.Itoa(fd))
	}

	reqBody := AppendArg{
		Fd: fd,
		Data: []byte(content),
	}
	respBody := AppendRet{}

	err = post("append", reqBody, &respBody)
	if err != nil {
		return 0, withStackedMessagef(err, "[%d] dfsAppend", fd)
	}

	bytesWritten = int64(respBody.BytesWritten)
	return bytesWritten, nil
}

func dfsScan(path string) (fileInfos []FileInfo, err error) {
	absPath := checkPath(path)
	if absPath == "" {
		return nil, withStackedMessagef(InvalidPathErr, "[%s] dfsScan", path)
	}

	reqBody := ScanArg{
		Path: absPath,
	}
	respBody := ScanRet{}

	err = post("scan", reqBody, &respBody)
	if err != nil {
		return nil, withStackedMessagef(err, "[%s] dfsScan", path)
	}

	split := strings.Split(path, "/")
	name := split[len(split)-1]

	return respFileInfos2FileInfos(name, respBody.FileInfos), nil
}

func dfsStat(path string) (fileInfo FileInfo, err error) {
	absPath := checkPath(path)
	if absPath == "" {
		return FileInfo{}, withStackedMessagef(InvalidPathErr, "[%s] dfsStat", path)
	}

	reqBody := GetFileInfoArg{
		Path: absPath,
	}
	respBody := GetFileInfoRet{}

	err = post("fileInfo", reqBody, &respBody)
	if err != nil {
		return FileInfo{}, withStackedMessagef(err, "[%s] dfsStat", path)
	}

	if !respBody.Exist {
		return FileInfo{}, withStackedMessagef(StatNonExistentErr, "[%s] dfsStat", path)
	} else {
		split := strings.Split(path, "/")
		name := split[len(split)-1]
		return FileInfo{
			Name: name,
			IsDir: respBody.IsDir,
		}, nil
	}
}


// helper functions
const (
	dfsRoot		=	""
)

var (
	clientAddr	=	"http://1.15.127.43:1333"
)

func post(api string, reqBody interface{}, respBody interface{}) (err error) {
	url := clientAddr + "/" + api
	reqBodyRaw, _ := json.Marshal(reqBody)

	logger.Debugf("[%s] Send Post: %s", url, reqBodyRaw)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyRaw))
	if err != nil {
		return err
	}

	respBodyRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.Debugf("[%s] Get Post Raw: %s", url, respBodyRaw)

	if respBody != nil {
		err = json.Unmarshal(respBodyRaw, respBody)
		if err != nil {
			return err
		}
	}

	logger.Debugf("[%s] Get Post Json Response: %v", url, respBody)

	return nil
}

func checkPath(path string) (absPath string) {
	if path[0] != '/' {
		return ""
	} else {
		return transPath(path)
	}
}

func transPath(relPath string) (absPath string) {
	return dfsRoot + relPath
}

func withStackedMessagef(before error, format string, args ...interface{}) (after error) {
	return errors.WithStack(errors.WithMessagef(before, format, args...))
}

func respFileInfos2FileInfos(name string, before []GetFileInfoRet) (after []FileInfo) {
	for i := 0; i < len(before); i += 1 {
		after = append(after, FileInfo{
			Name: name,
			IsDir: before[i].IsDir,
		})
	}

	return after
}