package main

import (
	"DFS/master"
	"DFS/util"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"testing"
)

func TestLogAndCheckpointSingle(t *testing.T){
	_,m,_ := initTest()
	//delete old ckp
	_, err := os.Stat("../log/checkpoint.dat")
	if err == nil {
		err := os.Remove("../log/checkpoint.dat")
		if err != nil {
			logrus.Fatalf("mkdir %v error\n", "cs")
		}
	}
	defer m.Exit()
	var createReply util.CreateRet
	var mkdirReply util.MkdirRet
	var listReply util.ListRet
	err = m.CreateRPC(util.CreateArg{Path: "/file1"}, &createReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/file2"}, &createReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/file3"}, &createReply)
	util.AssertNil(t,err)
	err = m.MkdirRPC(util.MkdirArg{Path: "/dir1"}, &mkdirReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file1"}, &createReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file2"}, &createReply)
	util.AssertNil(t,err)
	err = m.ListRPC(util.ListArg{Path: "/"}, &listReply)
	util.AssertNil(t,err)
	for index, file := range listReply.Files {
		util.AssertEqual(t,file,"file"+strconv.Itoa(index+1))
	}
	err = HTTPDelete(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	fd,err := HTTPOpen(util.CLIENTADDR,"/dir1/file2")
	util.AssertNil(t,err)
	// Write 3.5 chunks to /dir1/file2
	offset := 0
	data := []byte(util.MakeString(util.MAXCHUNKSIZE*3.5))
	err = HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)
	// Write 1 chunk at offset 3*size+1 in /dir1/file2
	offset = util.MAXCHUNKSIZE*3+1
	data = []byte(util.MakeString(util.MAXCHUNKSIZE-1))
	err = HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)

	// Read 65 bytes near the chunk 3
	result,err := HTTPRead(util.CLIENTADDR,fd,util.MAXCHUNKSIZE*3-1,util.MAXCHUNKSIZE+1)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"jkabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk")

	_,err = HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)

	// storeCheckPoint
	err = m.StoreCheckPoint()
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file3"}, &createReply)
	util.AssertNil(t,err)
	fd,err = HTTPOpen(util.CLIENTADDR,"/dir1/file3")
	util.AssertNil(t,err)

	// Write 3.5 chunks to /dir1/file3
	offset = 0
	data = []byte(util.MakeString(util.MAXCHUNKSIZE*3.5))
	err = HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)
	_,err = HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)
	// After all those operations, the namespace should be :
	/*
		/: file1 file2
		/dir1 : file1 file2(write sth in) file3(write after ckp)
	*/
	m.Exit()
	// restart master
	m,_ = master.InitMaster(util.MASTER1ADDR, "../log")
	go func(){m.Serve()}()

	// Read 65 bytes near the /dir1/file2 chunk 3
	fd,err = HTTPOpen(util.CLIENTADDR,"/dir1/file2")
	util.AssertNil(t,err)
	result,err = HTTPRead(util.CLIENTADDR,fd,util.MAXCHUNKSIZE*3-1,util.MAXCHUNKSIZE+1)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"jkabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk")
	_,err = HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)

	fd,err = HTTPOpen(util.CLIENTADDR,"/dir1/file3")
	util.AssertNil(t,err)
	result,err = HTTPRead(util.CLIENTADDR,fd,util.MAXCHUNKSIZE,20)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"mnopqrstuvwxyzabcdef")
	_,err = HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)
}
