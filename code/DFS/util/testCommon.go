package util

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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
		t.Fatal("AssertEqual Failed :", a, b)
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
	fd,err = strconv.Atoi(string(body))
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
// HTTPWrite : write a file according to fd
func HTTPWrite(addr string,fd int,offset int,data []byte)(err error){
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



// HTTPRead : read a file according to fd
func HTTPRead(addr string,fd int,offset int,len int)(result ReadRet,err error){
	url := "http://"+addr+"/read"
	postBody, _ := json.Marshal(map[string]interface{}{
		"fd":  fd,
		"offset" :offset,
		"Len" : len,
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
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&result)
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

