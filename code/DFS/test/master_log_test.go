package test

import (
	"DFS/chunkserver"
	"DFS/client"
	"DFS/master"
	"DFS/util"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"testing"
	"time"
)

// initTest init and return the handle of client,master and chunkservers
// TODO : remove duplication
func InitTest2() (c *client.Client,m *master.Master,cs []*chunkserver.ChunkServer){
	logrus.SetLevel(logrus.DebugLevel)
	//delete old ckp
	util.DeleteFile("../log/checkpoint.dat")
	//delete old log
	util.DeleteFile("../log/log.dat")
	// Init master and client
	m,_ = master.InitMaster(util.MASTER1ADDR, "../log")
	go func(){m.Serve()}()
	c = client.InitClient(util.CLIENTADDR, util.MASTER1ADDR)
	go func(){c.Serve()}()

	// Register some virtual chunkServers
	_, err := os.Stat("ck")
	if err == nil {
		err := os.RemoveAll("ck")
		if err != nil {
			logrus.Fatalf("mkdir %v error\n", "ck")
		}
	}
	err = os.Mkdir("ck", 0755)
	if err != nil {
		logrus.Fatalf("mkdir %v error\n", "ck")
	}
	cs = make([]*chunkserver.ChunkServer, 5)
	for index,port := range []int{3000,3001,3002,3003,3004}{
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		cs[index] = chunkserver.InitChunkServer(string(addr), "ck/ck"+strconv.Itoa(port),  util.MASTER1ADDR)
		_ = m.RegisterServer(addr)
		//util.AssertNil(t,err)
	}
	time.Sleep(500*time.Millisecond)
	return
}

//func TestLogMultiMaster(t *testing.T){
//	masterList := make([]*master.Master,3)
//	addrList := make([]string,3)
//	addrList[0]=util.MASTER1ADDR
//	addrList[1]=util.MASTER2ADDR
//	addrList[2]=util.MASTER3ADDR
//	for i:=0;i<3;i++{
//		go func(order int) {
//			m,err := master.InitMultiMaster(util.Address(addrList[order]), util.LinuxPath(os.Args[5+order]))
//			if err!=nil{
//				fmt.Println("Master Init Error :",err)
//				os.Exit(1)
//			}
//			m.Serve()
//		}(i)
//	}
//}
func TestLogAndCheckpointSingle(t *testing.T){
	c,m,cs := InitTest2()
	//delete old ckp
	util.DeleteFile("../log/checkpoint.dat")
	//delete old log
	util.DeleteFile("../log/log.dat")
	var createReply util.CreateRet
	var mkdirReply util.MkdirRet
	var listReply util.ListRet
	err := m.CreateRPC(util.CreateArg{Path: "/file1"}, &createReply)
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
	util.AssertEqual(t,4,len(listReply.Files))
	err = util.HTTPDelete(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	fd,err := util.HTTPOpen(util.CLIENTADDR,"/dir1/file2")
	util.AssertNil(t,err)
	// Write 3.5 chunks to /dir1/file2
	offset := 0
	data := []byte(util.MakeString(util.MAXCHUNKSIZE*3.5))
	err = util.HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)
	// Write 1 chunk at offset 3*size+1 in /dir1/file2
	offset = util.MAXCHUNKSIZE*3+1
	data = []byte(util.MakeString(util.MAXCHUNKSIZE-1))
	err = util.HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)

	// Read 65 bytes near the chunk 3
	result,err := util.HTTPRead(util.CLIENTADDR,fd,util.MAXCHUNKSIZE*3-1,util.MAXCHUNKSIZE+1)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"jkabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk")

	_,err = util.HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)

	// storeCheckPoint
	err = m.StoreCheckPoint()
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file3"}, &createReply)
	util.AssertNil(t,err)
	fd,err = util.HTTPOpen(util.CLIENTADDR,"/dir1/file3")
	util.AssertNil(t,err)

	// Write 3.5 chunks to /dir1/file3
	offset = 0
	data = []byte(util.MakeString(util.MAXCHUNKSIZE*3.5))
	err = util.HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)
	_,err = util.HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)
	// After all those operations, the namespace should be :
	/*
		/: file1 file2
		/dir1 : file1 file2(write sth in) file3(write after ckp)
	*/
	m.Exit()
	// restart master
	m,err = master.InitMaster(util.MASTER1ADDR, "../log")
	util.AssertNil(t,err)
	go func(){m.Serve()}()

	// Read 65 bytes near the /dir1/file2 chunk 3
	fd,err = util.HTTPOpen(util.CLIENTADDR,"/dir1/file2")
	util.AssertNil(t,err)
	result,err = util.HTTPRead(util.CLIENTADDR,fd,util.MAXCHUNKSIZE*3-1,util.MAXCHUNKSIZE+1)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"jkabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk")
	_,err = util.HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)

	fd,err = util.HTTPOpen(util.CLIENTADDR,"/dir1/file3")
	util.AssertNil(t,err)
	result,err = util.HTTPRead(util.CLIENTADDR,fd,util.MAXCHUNKSIZE,20)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"mnopqrstuvwxyzabcdef")
	_,err = util.HTTPClose(util.CLIENTADDR,fd)
	util.AssertNil(t,err)

	// list /
	err = m.ListRPC(util.ListArg{Path: "/"}, &listReply)
	util.AssertNil(t,err)
	util.AssertEqual(t,3,len(listReply.Files))
	m.Exit()
	c.Exit()
	for _,_cs := range cs{
		_cs.Exit()
	}
}
