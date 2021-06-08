package main

import (
	"DFS/client"
	"DFS/master"
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
	WriteTest(c,m)
}
func WriteTest(c *client.Client,m *master.Master){
	fd,err := HTTPOpen(CLIENTADDR,"/file1")
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

	offset := 0
	data := make([]byte,256)
	err = HTTPWrite(CLIENTADDR,fd,offset,data)
	if err!=nil{
		fmt.Println(err)
	}
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
