package dfs

import (
	"backend/utils/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
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
func dfsOpen(path string, _ bool) (fd int, err error) {
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

	err = dfsNameX(path[:strings.LastIndex(path, "/")])
	if err != nil {
		return 0, errors.WithMessagef(err, "[%s] dfsCreate")
	}

	// create file
	createReqBody := CreateArg{
		Path: transPath(path),
	}

	err = post("create", createReqBody, nil)
	if err != nil {
		return 0, withStackedMessagef(err, "[%s] dfsCreate cannot create leaf file", path)
	}

	fd, err = dfsOpen(path, false)
	if err != nil {
		return 0, withStackedMessagef(err, "[%s] dfsCreate cannot open leaf file", path)
	}

	return fd, nil
}

func dfsMkdir(path string) (err error) {
	err = dfsNameX(path)
	if err != nil {
		return errors.WithMessagef(err, "[%s] dfsMkdir")
	}

	return nil
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

func dfsRead(fd int, off int64, len int64) (data []byte, err error) {
	if fd <= 0 {
		return nil, withStackedMessagef(InvalidFdErr, "[%d] dfsRead", strconv.Itoa(fd))
	}

	reqBody := ReadArg{
		Fd: fd,
		Offset: int(off),
		Len: int(len),
	}

	var respBody []byte
	err = post("read", reqBody, &respBody, true)
	if err != nil {
		return nil, withStackedMessagef(err, "[%d] dfsRead", fd)
	}

	data = filterAllPadding(respBody)
	return data, nil
}

func dfsReadAll(path string) (data []byte, err error) {
	absPath := checkPath(path)
	if absPath == "" {
		return nil, withStackedMessagef(InvalidPathErr, "[%s] dfsReadAll", path)
	}

	reqBody := GetFileInfoArg{
		Path: absPath,
	}
	respBody := GetFileInfoRet{}

	err = post("fileInfo", reqBody, &respBody)
	if err != nil {
		return nil, withStackedMessagef(err, "[%s] dfsReadAll", path)
	}

	if !respBody.Exist {
		return nil, withStackedMessagef(StatNonExistentErr, "[%s] dfsReadAll", path)
	} else {
		sup := int64(respBody.UpperFileSize)
		fd, err := dfsOpen(path, false)
		if err != nil {
			return nil, withStackedMessagef(err, "[%s] dfsReadAll fail to open", path)
		}

		data, err := dfsRead(fd, 0, sup)
		if err != nil {
			return nil, withStackedMessagef(err, "[%s] dfsReadAll fail to read", path)
		}

		return data, nil
	}
}

func dfsWrite(fd int, off int64, data []byte) (bytesWritten int64, err error) {
	if fd <= 0 {
		return 0, withStackedMessagef(InvalidFdErr, "[%d] dfsWrite", strconv.Itoa(fd))
	}

	reqBody := map[string]string {
		"fd": strconv.Itoa(fd),
		"offset": strconv.FormatInt(off, 10),
	}
	respBody := WriteRet{}

	//err = post("write", reqBody, &respBody)
	err = postForm("write", data, reqBody, &respBody)
	if err != nil {
		return 0, withStackedMessagef(err, "[%d] dfsWrite", fd)
	}

	bytesWritten = int64(respBody.BytesWritten)
	return bytesWritten, nil
}

func dfsAppend(fd int, data []byte) (bytesWritten int64, err error) {
	if fd <= 0 {
		return 0, withStackedMessagef(InvalidFdErr, "[%d] dfsSppend", strconv.Itoa(fd))
	}

	reqBody := map[string]string {
		"fd": strconv.Itoa(fd),
	}
	respBody := AppendRet{}

	//err = post("append", reqBody, &respBody)
	err = postForm("write", data, reqBody, &respBody)
	if err != nil {
		return 0, withStackedMessagef(err, "[%d] dfsAppend", fd)
	}

	bytesWritten = int64(len(data))
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

func post(api string, reqBody interface{}, respBody interface{}, returnRaw ...bool) (err error) {
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

	if len(returnRaw) > 0 && returnRaw[0] {
		*respBody.(*[]byte) = respBodyRaw
		return nil
	}

	if respBody != nil {
		err = json.Unmarshal(respBodyRaw, respBody)
		if err != nil {
			return err
		}
	}

	logger.Debugf("[%s] Get Post Json Response: %v", url, respBody)

	return nil
}

func postForm(api string, data []byte, params map[string]string, respBody interface{}) (err error) {
	url := clientAddr + "/" + api

	logger.Debugf("[%s] Send Post: %+v", url, params)

	bodyBuf := bytes.Buffer{}
	bodyWrite := multipart.NewWriter(&bodyBuf)
	file, err := bodyWrite.CreateFormFile("file", "raw")
	if err != nil {
		return err
	}

	err = ioWriteAll(file, data)
	if err != nil {
		return err
	}

	for k, v := range params {
		field, err := bodyWrite.CreateFormField(k)
		if err != nil {
			return err
		}

		err = ioWriteAll(field, []byte(v))
		if err != nil {
			return err
		}
	}

	err = bodyWrite.Close()
	if err != nil {
		return err
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, &bodyBuf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", bodyWrite.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBodyRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.Debugf("[%s] Get PostForm Raw: %s", url, respBodyRaw)

	if respBody != nil {
		err = json.Unmarshal(respBodyRaw, respBody)
		if err != nil {
			return err
		}
	}

	logger.Debugf("[%s] Get PostForm Json Response: %v", url, respBody)

	return nil

}

func ioWriteAll(writer io.Writer, data []byte) (err error) {
	written, total := 0, len(data)
	for written < total {
		n, err := writer.Write(data[written:])
		if err != nil {
			return err
		}

		written += n
	}

	if written != total {
		return fmt.Errorf("in postForm, expect to write %d bytes, actually it is %d", total, written)
	}

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
			Name: before[i].FileName,
			IsDir: before[i].IsDir,
		})
	}

	return after
}

func filterAllPadding(before []byte) (after []byte) {
	i := 0
	for _, val := range before {
		if val != 0 {
			before[i] = val
			i += 1
		}
	}

	return before[:i]
}

func dfsNameX(path string) (err error) {
	levels := strings.Split(path[1:], "/")
	curPath := ""
	for _, l := range levels {
		curPath += "/" + l
		statReqBody := GetFileInfoArg{
			Path: transPath(curPath),
		}
		statRespBody := GetFileInfoRet{}

		err = post("fileInfo", statReqBody, &statRespBody)
		if err != nil {
			return withStackedMessagef(err, "[%s] dfsNameX cannot stat parents \"%s\"", path, curPath)
		}

		if statRespBody.Exist {
			if !statRespBody.IsDir {
				return withStackedMessagef(err, "[%s] dfsNameX parent \"%s\" is file", path, curPath)
			} else {
				continue
			}
		} else {
			mkdirReqBody := MkdirArg{
				Path: transPath(curPath),
			}

			err = post("mkdir", mkdirReqBody, nil)
			if err != nil {
				return withStackedMessagef(err, "[%s] dfsNameX cannot create parents", path)
			}
		}
	}

	return nil
}