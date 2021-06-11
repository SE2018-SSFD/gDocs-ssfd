package main

import (
	"DFS/chunkserver"
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
	"testing"
	"time"
)


// TODO : use go testing package to rewrite an assert-style program
func initTest() (c *client.Client,m *master.Master,cs []*chunkserver.ChunkServer){
	logrus.SetLevel(logrus.DebugLevel)
	// Init master and client
	m = master.InitMaster(util.MASTERADDR, ".")
	go func(){m.Serve()}()
	c = client.InitClient(util.CLIENTADDR, util.MASTERADDR)
	go func(){c.Serve()}()
	// Register some virtual chunkServers
	cs = make([]*chunkserver.ChunkServer, 5)
	for index,port := range []int{3000,3001,3002,3003,3004}{
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		cs[index] = chunkserver.InitChunkServer(string(addr), "ck"+strconv.Itoa(port),  util.MASTERADDR)
		_ = m.RegisterServer(addr)
		//util.AssertNil(t,err)
	}
	time.Sleep(time.Second)
	return
}


func TestReadWrite(t *testing.T) {
	c,m,cs := initTest()
	err := HTTPCreate(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	err = HTTPCreate(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)
	fd1,err := HTTPOpen(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	fd2,err := HTTPOpen(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)

	// Write 4 chunks to file1
	offset := 0
	data := make([]byte,util.MAXCHUNKSIZE*4)
	err = HTTPWrite(util.CLIENTADDR,fd1,offset,data)
	util.AssertNil(t,err)
	fileState,err := HTTPGetFileInfo(util.CLIENTADDR,"/file1")
	fmt.Println(fileState)

	// Write 3.5 chunks to file2
	offset = 0
	data = make([]byte,util.MAXCHUNKSIZE*3.5)
	err = HTTPWrite(util.CLIENTADDR,fd2,offset,data)
	util.AssertNil(t,err)
	fileState,err = HTTPGetFileInfo(util.CLIENTADDR,"/file2")
	fmt.Println(fileState)

	// Write 1 chunk at offset 3 in file2
	offset = util.MAXCHUNKSIZE*3
	data = make([]byte,util.MAXCHUNKSIZE)
	err = HTTPWrite(util.CLIENTADDR,fd2,offset,data)
	util.AssertNil(t,err)
	fileState,err = HTTPGetFileInfo(util.CLIENTADDR,"/file2")
	fmt.Println(fileState)
	m.Exit()
	c.Exit()
}


func TestOpenClose(t *testing.T) {
	c,m,cs := initTest()
	err := HTTPCreate(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	err = HTTPCreate(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)
	fd,err := HTTPOpen(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	logrus.Infoln("fd :",fd)
	code,err := HTTPClose(util.CLIENTADDR,fd)
	util.AssertEqual(t,code,200)
	fd,err = HTTPOpen(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)
	logrus.Infoln("fd :",fd)
	code,err = HTTPClose(util.CLIENTADDR,fd)
	util.AssertEqual(t,code,200)
	util.AssertNil(t,err)
	code,err = HTTPClose(util.CLIENTADDR,fd)
	util.AssertEqual(t,code,400)
	m.Exit()
	c.Exit()
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
