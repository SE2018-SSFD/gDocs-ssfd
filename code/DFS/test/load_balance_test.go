package test

import (
	"DFS/chunkserver"
	"DFS/client"
	"DFS/master"
	"DFS/util"
	"DFS/util/zkWrap"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// initTest init and return the handle of client,master and chunkservers
// TODO : remove duplication
func initTestSingleLB() (c *client.Client,m *master.Master,cs []*chunkserver.ChunkServer){
	prepareLB()
	// start client and master
	m,_ = master.InitMaster(util.MASTER1ADDR, "../log")
	go func(){m.Serve()}()
	c = client.InitClient(util.CLIENTADDR, util.MASTER1ADDR)
	go func(){c.Serve()}()

	// Register some virtual chunkServers
	cs = make([]*chunkserver.ChunkServer, 5)
	for index,port := range []int{3000,3001,3002,3003,3004}{
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		cs[index] = chunkserver.InitChunkServer(string(addr), "ck/ck"+strconv.Itoa(port),  util.MASTER1ADDR)
		_ = m.RegisterServer(addr)
	}
	time.Sleep(500*time.Millisecond)
	return
}
func initTestMultiLB() (c *client.Client,masterList []*master.Master,cs []*chunkserver.ChunkServer){
	multiPrepareLB()
	// start client
	c = client.InitClient(util.CLIENTADDR, util.MASTER1ADDR)
	go func(){c.Serve()}()

	// start multiple master
	masterList = make([]*master.Master,3)
	addrList := make([]string,3)
	addrList[0]=util.MASTER1ADDR
	addrList[1]=util.MASTER2ADDR
	addrList[2]=util.MASTER3ADDR
	var wg sync.WaitGroup
	wg.Add(util.MASTERCOUNT)
	for i:=0;i<util.MASTERCOUNT;i++{
		go func(order int) {
			var mp *master.Master
			mp,err := master.InitMultiMaster(util.Address(addrList[order]), util.LinuxPath("log_"+strconv.Itoa(order+1)))
			if err!=nil{
				fmt.Println("Master Init Error :",err)
				os.Exit(1)
			}
			masterList[order] = mp
			wg.Done()
			mp.Serve()
		}(i)
	}
	wg.Wait()
	logrus.Debugln(masterList[0].GetStatusString())

	// Register some virtual chunkServers
	cs = make([]*chunkserver.ChunkServer, 5)
	for index,port := range []int{3000,3001,3002,3003,3004}{
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		cs[index] = chunkserver.InitChunkServer(string(addr), "ck/ck"+strconv.Itoa(port),  util.MASTER1ADDR)
	}

	return c,masterList,cs
}

func prepareLB(){
	logrus.SetLevel(logrus.DebugLevel)
	//delete old ckp
	util.DeleteFile("../log/checkpoint.dat")
	//delete old log
	util.DeleteFile("../log/log.dat")
	// Init master and client
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
}
func multiPrepareLB(){
	logrus.SetLevel(logrus.DebugLevel)
	//delete old ckp
	util.DeleteFile("../log/checkpoint.dat")
	//delete old log
	util.DeleteFile("../log/log.dat")
	_, err := os.Stat("ck")
	if err == nil {
		err := os.RemoveAll("ck")
		if err != nil {
			logrus.Fatalf("remove %v error\n", "ck")
		}
	}
	err = os.Mkdir("ck", 0755)
	if err != nil {
		logrus.Fatalf("mkdir %v error\n", "ck")
	}
	for i:=0;i<3;i++{
		filename := "log_"+strconv.Itoa(i+1)
		_, err = os.Stat(filename)
		if err == nil {
			logrus.Warnf("delete %v",filename)

			err = os.RemoveAll(filename)
			if err != nil {
				logrus.Fatalf("remove %v error\n", filename)
			}
		}
		err = os.Mkdir(filename, 0755)
		if err != nil {
			logrus.Fatalf("mkdir %v error\n", filename)
		}
	}
	err = zkWrap.Chroot("/DFS")
	if err!=nil{
		os.Exit(1)
	}
}
func TestLoadBalanceAllocate(t *testing.T){
	c,m,cs := initTestSingleLB()
	defer func() {
		m.Exit()
		c.Exit()
		for _,_cs := range cs{
			_cs.Exit()
		}
	}()
	var createReply util.CreateRet
	var mkdirReply util.MkdirRet
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
	fd,err := util.HTTPOpen(util.CLIENTADDR,"/dir1/file2")
	util.AssertNil(t,err)
	// Write 50 chunks to /dir1/file2
	offset := 0
	data := []byte(util.MakeString(util.MAXCHUNKSIZE*50))
	err = util.HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)
	result := m.GetServersChunkNum()
	for i:=0;i<5;i++{
		util.AssertEqual(t,result[i].ChunkNum,50 * 3 / 5)
	}
	offset = 50 * util.MAXCHUNKSIZE
	err = util.HTTPWrite(util.CLIENTADDR,fd,offset,data)
	result = m.GetServersChunkNum()
	for i:=0;i<5;i++{
		util.AssertEqual(t,result[i].ChunkNum,100 * 3 / 5)
	}
}

func TestLoadBalanceReallocate(t *testing.T){
	c,mList,csList:=initTestMultiLB()
	time.Sleep(1*util.HERETRYTIMES*time.Second)
	defer func() {
		for i:=0;i<util.MASTERCOUNT;i++{
			mList[i].Exit()
		}
		for i:=0;i<len(csList);i++{
			csList[i].Exit()
		}
		c.Exit()
	}()
	err := util.HTTPCreate(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	err = util.HTTPCreate(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)
	err = util.HTTPMkdir(util.CLIENTADDR,"/dir1")
	util.AssertNil(t,err)
	err = util.HTTPCreate(util.CLIENTADDR,"/dir1/file1")
	util.AssertNil(t,err)
	err = util.HTTPCreate(util.CLIENTADDR,"/dir1/file2")
	util.AssertNil(t,err)
	err = util.HTTPCreate(util.CLIENTADDR,"/file3")
	util.AssertNil(t,err)
	fd,err := util.HTTPOpen(util.CLIENTADDR,"/dir1/file2")
	util.AssertNil(t,err)

	leader := -1
	for i:=0;i<3;i++{
		if mList[i].IsLeader(){
			leader = i
			break
		}
	}
	// Write 50 chunks to /dir1/file2
	offset := 0
	data := []byte(util.MakeString(util.MAXCHUNKSIZE*50))
	err = util.HTTPWrite(util.CLIENTADDR,fd,offset,data)
	util.AssertNil(t,err)

	result := mList[leader].GetServersChunkNum()
	for i:=0;i<5;i++{
		util.AssertEqual(t,result[i].ChunkNum,50 * 3 / 5)
	}
	offset = 50 * util.MAXCHUNKSIZE
	err = util.HTTPWrite(util.CLIENTADDR,fd,offset,data)
	result =  mList[leader].GetServersChunkNum()
	for i:=0;i<5;i++{
		util.AssertEqual(t,result[i].ChunkNum,100 * 3 / 5)
	}
	addr := util.Address("127.0.0.1:" + strconv.Itoa(3005))
	_ = chunkserver.InitChunkServer(string(addr), "ck/ck"+strconv.Itoa(3005),  util.MASTER1ADDR)
	time.Sleep(time.Second)
	err = mList[leader].LoadBalanceCheck()
	util.AssertNil(t,err)
	result = mList[leader].GetServersChunkNum()
	for i:=0;i<5;i++{
		util.AssertGreater(t,result[i].ChunkNum,32) // 按目前的算法和测试参数，这里就是大于33

	}
}