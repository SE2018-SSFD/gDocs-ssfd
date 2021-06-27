package main
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