package main

import (
	"DFS/util"
	"fmt"
	"testing"
)

func TestLogAndCheckpointSingle(t *testing.T){
	m := initMasterTest()
	defer m.Exit()
	var createReply util.CreateRet
	var mkdirReply util.MkdirRet
	var listReply util.ListRet
	var deleteReply util.DeleteRet
	var getFileMetaReply util.GetFileMetaRet
	err := m.CreateRPC(util.CreateArg{Path: "/file1"}, &createReply)
	util.AssertNil(t,err)
	err := m.CreateRPC(util.CreateArg{Path: "/file2"}, &createReply)
	util.AssertNil(t,err)
	err := m.CreateRPC(util.CreateArg{Path: "/file3"}, &createReply)
	util.AssertNil(t,err)
	err = m.MkdirRPC(util.MkdirArg{Path: "/dir1"}, &mkdirReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file1"}, &createReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file2"}, &createReply)
	util.AssertNil(t,err)
	err = m.ListRPC(util.ListArg{Path: "/dir1"}, &listReply)
	util.AssertNil(t,err)
	for _, file := range listReply.Files {
		fmt.Print(file, " ")
	}
}
//
//import (
//	"DFS/chunkserver"
//	"DFS/client"
//	"DFS/master"
//	"DFS/util"
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"github.com/sirupsen/logrus"
//	"io/ioutil"
//	"net/http"
//	"strconv"
//	"testing"
//	"time"
//)
//
//
//// initTest init and return the handle of client,master and chunkservers
//func initTest() (c *client.Client,m *master.Master,cs []*chunkserver.ChunkServer){
//	logrus.SetLevel(logrus.DebugLevel)
//
//	// Init master and client
//	m,err := master.InitMaster(util.MASTER1ADDR, ".")
//	go func(){m.Serve()}()
//	c = client.InitClient(util.CLIENTADDR, util.MASTER1ADDR)
//	go func(){c.Serve()}()
//
//	// Register some virtual chunkServers
//	cs = make([]*chunkserver.ChunkServer, 5)
//	for index,port := range []int{3000,3001,3002,3003,3004}{
//		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
//		cs[index] = chunkserver.InitChunkServer(string(addr), "ck"+strconv.Itoa(port),  util.MASTER1ADDR)
//		_ = m.RegisterServer(addr)
//		//util.AssertNil(t,err)
//	}
//
//	time.Sleep(time.Second)
//	return
//}