package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"testing"
)

const (
	CLIENTADDR     = "127.0.0.1:1233"
	MASTER1ADDR     = "127.0.0.1:1234"
	MASTER2ADDR     = "127.0.0.1:1235"
	MASTER3ADDR     = "127.0.0.1:1236"

	PARALLELSIZE   = 3
	MAXWAITINGTIME = 5 // the time to wait for parallel tests finished
)

// Assert if a string contains same data
func AssertSameData(t *testing.T,a []byte){
	if len(a) > 0{
		r := a[0]
		for _,char := range a{
			if char!=r{
				t.Fail()
				t.Fatal("AssertSameData failed : ",r,char)
			}
		}
	}
	t.Logf("AssertSameData Succeed : ")

}

func AssertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Fail()
		t.Logf("AssertEqual Failed %v,%v", a, b)
		return
	}
	t.Logf("AssertEqual Succeed :")

}
func AssertGreater(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Fail()
		t.Fatal("AssertGreater Failed :", a, b)
	}
}
func AssertTrue(t *testing.T, a bool) {
	t.Helper()
	if a != true {
		t.Fail()
		t.Fatal("AssertTrue Failed :", a)
	}
	t.Logf("AssertTrue Succeed :")

}
func AssertNotTrue(t *testing.T, a bool) {
	t.Helper()
	if a != false {
		t.Fail()
		t.Fatal("AssertNotTrue Failed :", a)
	}
	t.Logf("AssertNotTrue Succeed :")

}
func DeleteFile(path string){
	_, err := os.Stat(path)
	if err == nil {
		err := os.Remove(path)
		if err != nil {
			logrus.Fatalf("delete file failed")
		}
	}
}
func AssertNotEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a == b {
		t.Fail()
		t.Fatal("AssertNotEqual Failed :", a, b)
	}
	t.Logf("AssertNotEqual Succeed :")
}
func AssertNil(t *testing.T, a interface{}) {
	t.Helper()
	if a != nil {
		t.Fail()
		t.Fatal("AssertNil Failed :", a)
	}
	t.Logf("AssertNil Succeed :")
}
func AssertNotNil(t *testing.T, a interface{}) {
	t.Helper()
	if a == nil {
		t.Fail()
		t.Fatal("AssertNotNil Failed :", a)
	}
	t.Logf("AssertNotNil Succeed :")

}
// HTTPOpen : open a file
// return file's fd on success
func HTTPOpen(addr string,path string)(fd int,err error){
	url := "http://"+addr+"/open"
	postBody, _ := json.Marshal(map[string]string{
		"path":  path,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var ret OpenRet
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&ret)
	fd = ret.Fd
	return
}

// HTTPClose : close a file according to fd
func HTTPClose(addr string,fd int)(statusCode int,err error){
	url := "http://"+addr+"/close"
	postBody, _ := json.Marshal(map[string]interface{}{
		"fd":  fd,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	statusCode = resp.StatusCode
	defer resp.Body.Close()
	return
}

// HTTPCreate : crate a file
func HTTPCreate(addr string,path string)(err error){
	url := "http://"+addr+"/create"
	postBody, _ := json.Marshal(map[string]string{
		"path":  path,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return
}

// HTTPDelete : delete a file
func HTTPDelete(addr string,path string)(err error){
	url := "http://"+addr+"/delete"
	postBody, _ := json.Marshal(map[string]string{
		"path":  path,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return
}
// HTTPWriteDeprecated : write a file according to fd
func HTTPWriteDeprecated(addr string,fd int,offset int,data []byte)(err error){
	url := "http://"+addr+"/write"
	postBody, _ := json.Marshal(map[string]interface{}{
		"fd":  fd,
		"offset" :offset,
		"data" : data,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return
}

// HTTPWrite : write a file according to fd
func HTTPWrite(addr string,fd int,offset int,data []byte)(err error){
	url := "http://"+addr+"/write"
	bodyBuf := bytes.Buffer{}
	bodyWrite := multipart.NewWriter(&bodyBuf)
	file, err := bodyWrite.CreateFormFile("file", "raw")
	err = ioWriteAll(file, data)

	if err != nil {
		return err
	}
	params := make(map[string]string)
	params["fd"] = strconv.Itoa(fd)
	params["offset"] = strconv.Itoa(offset)

	for k, v := range params {
		field, errr := bodyWrite.CreateFormField(k)
		if errr != nil {
			return errr
		}
		errr = ioWriteAll(field, []byte(v))
		if errr != nil {
			return errr
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
	logrus.Warnf(url)
	resp, err := client.Do(req)
	logrus.Warnln(err)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return
}


// HTTPAppendDeprecated : append a file according to fd
func HTTPAppendDeprecated(addr string,fd int,data []byte)(result CAppendRet,err error){
	url := "http://"+addr+"/append"
	postBody, _ := json.Marshal(map[string]interface{}{
		"fd":  fd,
		"data" : data,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&result)
	return
}
// HTTPAppend : append a file according to fd
func HTTPAppend(addr string,fd int,data []byte)(result CAppendRet,err error){
	url := "http://"+addr+"/append"
	bodyBuf := bytes.Buffer{}
	bodyWrite := multipart.NewWriter(&bodyBuf)
	file, err := bodyWrite.CreateFormFile("file", "raw")
	err = ioWriteAll(file, data)
	if err != nil {
		return
	}
	params := make(map[string]string)
	params["fd"] = strconv.Itoa(fd)

	for k, v := range params {
		field, errr := bodyWrite.CreateFormField(k)
		if errr != nil {
			return
		}

		errr = ioWriteAll(field, []byte(v))
		if errr != nil {
			return
		}
	}
	err = bodyWrite.Close()
	if err != nil {
		return
	}
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, &bodyBuf)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", bodyWrite.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&result)
	return
}


// HTTPRead : read a file according to fd
func HTTPRead(addr string,fd int,offset int,length int)(result ReadRet,err error){
	url := "http://"+addr+"/read"
	postBody, _ := json.Marshal(map[string]interface{}{
		"fd":  fd,
		"offset" :offset,
		"Len" : length,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}

	respBodyRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	result.Data = respBodyRaw
	result.Len = len(respBodyRaw)
	return
}

// HTTPGetFileInfo : get file info according to path
func HTTPGetFileInfo(addr string, path string) (fileState GetFileMetaRet, err error) {
	url := "http://"+addr+"/fileInfo"
	postBody, _ := json.Marshal(map[string]interface{}{
		"path":  path,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&fileState)
	return
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