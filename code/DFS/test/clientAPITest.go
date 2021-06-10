package main

import (
	"DFS/client"
	"DFS/master"
	"DFS/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	CLIENTADDR = "127.0.0.1:1233"
	MASTERADDR = "127.0.0.1:1234"

)
// TODO : use go testing package to rewrite an assert-style program
func main() {
	logrus.SetLevel(logrus.DebugLevel)

	// Init master and client
	m := master.InitMaster(MASTERADDR, ".")
	go func(){m.Serve()}()
	c := client.InitClient(CLIENTADDR, MASTERADDR)
	go func(){c.Serve()}()
	time.Sleep(time.Second)

	// Start testing
	OpenCloseTest(c,m)
	ReadWriteTest(c,m)
}
func ReadWriteTest(c *client.Client,m *master.Master){
	fd1,err := HTTPOpen(CLIENTADDR,"/file1")
	if err != nil {
		fmt.Println(err)
		return
	}
	fd2,err := HTTPOpen(CLIENTADDR,"/file2")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Register some virtual chunkServers
	err = m.RegisterServer("127.0.0.1:3000")
	err = m.RegisterServer("127.0.0.1:3001")
	err = m.RegisterServer("127.0.0.1:3002")
	err = m.RegisterServer("127.0.0.1:3003")
	err = m.RegisterServer("127.0.0.1:3004")

	// Write 4 chunks to file1
	offset := 0
	data := make([]byte,util.MAXCHUNKSIZE*4)
	err = HTTPWrite(CLIENTADDR,fd1,offset,data)
	if err!=nil{
		fmt.Println(err)
	}
	fileState,err := HTTPGetFileInfo(CLIENTADDR,"/file1")
	fmt.Println(fileState)

	// Write 3.5 chunks to file2
	offset = 0
	data = make([]byte,util.MAXCHUNKSIZE*3.5)
	err = HTTPWrite(CLIENTADDR,fd2,offset,data)
	if err!=nil{
		fmt.Println(err)
	}
	fileState,err = HTTPGetFileInfo(CLIENTADDR,"/file2")
	fmt.Println(fileState)

	// Write 1 chunk at offset 3 in file2
	offset = util.MAXCHUNKSIZE*3
	data = make([]byte,util.MAXCHUNKSIZE)
	err = HTTPWrite(CLIENTADDR,fd2,offset,data)
	if err!=nil{
		fmt.Println(err)
	}
	fileState,err = HTTPGetFileInfo(CLIENTADDR,"/file2")
	fmt.Println(fileState)
}


func OpenCloseTest(c *client.Client,m *master.Master) {
	err := HTTPCreate(CLIENTADDR,"/file1")
	if err != nil {
		fmt.Println(err)
	}
	err = HTTPCreate(CLIENTADDR,"/file2")
	if err != nil {
		fmt.Println(err)
	}
	fd,err := HTTPOpen(CLIENTADDR,"/file1")
	if err != nil {
		fmt.Println(err)
	}else{
		logrus.Infoln("fd :",fd)
	}
	err = HTTPClose(CLIENTADDR,fd)
	if err != nil {
		fmt.Println(err)
	}
	fd,err = HTTPOpen(CLIENTADDR,"/file2")
	if err != nil {
		fmt.Println(err)
	}else{
		logrus.Infoln("fd :",fd)
	}
	err = HTTPClose(CLIENTADDR,fd)
	if err != nil {
		fmt.Println(err)
	}
	err = HTTPClose(CLIENTADDR,fd)
	if err != nil {
		fmt.Println(err)
	}
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
func HTTPClose(addr string,fd int)(err error){
	url := "http://"+addr+"/close"
	postBody, _ := json.Marshal(map[string]interface{}{
		"fd":  fd,
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

// HTTPGetFileInfo : get file info according to path
func HTTPGetFileInfo(addr string, path string) (fileState util.GetFileMetaRet, err error) {
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
