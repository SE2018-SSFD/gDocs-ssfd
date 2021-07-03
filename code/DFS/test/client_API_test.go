package test

import (
	"DFS/chunkserver"
	"DFS/client"
	"DFS/master"
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)
// initTest init and return the handle of client,master and chunkservers
func InitTest() (c *client.Client,m *master.Master,csList []*chunkserver.ChunkServer){
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
	csList = make([]*chunkserver.ChunkServer, 5)
	for index,port := range []int{3000,3001,3002,3003,3004}{
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		csList[index] = chunkserver.InitChunkServer(string(addr), "ck/ck"+strconv.Itoa(port),  util.MASTER1ADDR)
		_ = m.RegisterServer(addr)
		//util.AssertNil(t,err)
	}
	time.Sleep(500*time.Millisecond)
	return
}

// initTest init and return the handle of client,master and chunkservers
func InitMultiTest() (cList []*client.Client,m *master.Master,csList []*chunkserver.ChunkServer){
	logrus.SetLevel(logrus.DebugLevel)
	//delete old ckp
	util.DeleteFile("../log/checkpoint.dat")
	//delete old log
	util.DeleteFile("../log/log.dat")
	// Init master and client
	m,_ = master.InitMaster(util.MASTER1ADDR, "../log")
	go func(){m.Serve()}()
	//Init four clients
	cList = make([]*client.Client, 4)
	for index, port := range []int{1333, 1334, 1335, 1336} {
		go func(order int, p int) {
			addr := util.Address("127.0.0.1:" + strconv.Itoa(p))
			cList[order] = client.InitClient(addr, util.MASTER1ADDR)
			logrus.Infof("Init Client %v Success",addr)
			cList[order].Serve()
		}(index, port)
	}

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
	csList = make([]*chunkserver.ChunkServer, 5)
	for index,port := range []int{3000,3001,3002,3003,3004}{
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		csList[index] = chunkserver.InitChunkServer(string(addr), "ck/ck"+strconv.Itoa(port),  util.MASTER1ADDR)
		_ = m.RegisterServer(addr)

		//util.AssertNil(t,err)
	}
	time.Sleep(1000*time.Millisecond)
	return
}


// Test single-client read & write operation
func TestReadWriteSingle(t *testing.T) {
	c,m,csList := InitTest()
	defer func() {
		m.Exit()
		c.Exit()
		for _,_cs := range csList{
			_cs.Exit()
		}
	}()
	err := util.HTTPCreate(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	err = util.HTTPCreate(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)
	fd1,err := util.HTTPOpen(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	fd2,err := util.HTTPOpen(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)

	// Write 4 chunks to file1
	offset := 0
	data := make([]byte,util.MAXCHUNKSIZE*4)
	err = util.HTTPWrite(util.CLIENTADDR,fd1,offset,data)
	util.AssertNil(t,err)
	fileState,err := util.HTTPGetFileInfo(util.CLIENTADDR,"/file1")
	fmt.Println(fileState)

	// Write 3.5 chunks to file2
	offset = 0
	data = []byte(util.MakeString(util.MAXCHUNKSIZE*3.5))
	err = util.HTTPWrite(util.CLIENTADDR,fd2,offset,data)
	util.AssertNil(t,err)
	fileState,err = util.HTTPGetFileInfo(util.CLIENTADDR,"/file2")
	fmt.Println(fileState)

	// Write 1 chunk at offset 3*size+1 in file2
	offset = util.MAXCHUNKSIZE*3+1
	data = []byte(util.MakeString(util.MAXCHUNKSIZE-1))
	err = util.HTTPWrite(util.CLIENTADDR,fd2,offset,data)
	util.AssertNil(t,err)
	fileState,err = util.HTTPGetFileInfo(util.CLIENTADDR,"/file2")
	fmt.Println(fileState)

	// Read 65 bytes near the chunk 3
	result,err := util.HTTPRead(util.CLIENTADDR,fd2,util.MAXCHUNKSIZE*3-1,util.MAXCHUNKSIZE+1)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"jkabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk")

	// Read 66 bytes near the chunk 2
	result,err = util.HTTPRead(util.CLIENTADDR,fd2,util.MAXCHUNKSIZE*2-1,util.MAXCHUNKSIZE+2)
	util.AssertNil(t,err)
	util.AssertEqual(t,string(result.Data),"xyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk")
}

// Test multiple clients read & write operations with multiple masters
func TestReadWriteMulti(t *testing.T){
	cList,m,csList := InitMultiTest()
	defer func() {
		m.Exit()
		for _,c := range cList{
			c.Exit()
		}
		for _,cs := range csList{
			cs.Exit()
		}
	}()
	// to save fd in every client
	fdList := make([][]int,4)
	for index,_ := range fdList{
		fdList[index] = make([]int,4)
	}
	// create 4 testing files
	for i:=0;i<4;i++{
		path := "/file" + strconv.Itoa(i+1)
		err := util.HTTPCreate(string(cList[0].GetClientAddr()), path)
		util.AssertNil(t, err)
	}
	// open all files in all clients
	for i:=0;i<4;i++{
		clientAddr:= cList[i].GetClientAddr()
		for j:=0;j<4;j++ {
			var err error
			path := "/file" + strconv.Itoa(j+1)
			fdList[i][j], err = util.HTTPOpen(string(clientAddr), path)
			util.AssertNil(t, err)
		}
	}

	// 4 clients write to different file
	var wg sync.WaitGroup
	wg.Add(4)
	for i:=0;i<4;i++ {
		go func(index int) {
			clientAddr:= cList[index].GetClientAddr()
			offset := 0
			data := []byte(util.MakeInt(index,2*util.MAXCHUNKSIZE))
			err := util.HTTPWrite(string(clientAddr),fdList[index][index],offset, data)
			util.AssertNil(t,err)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i:=0;i<4;i++ {
		offset := 0
		result,err := util.HTTPRead(string(cList[0].GetClientAddr()),fdList[0][i],offset,2*util.MAXCHUNKSIZE)
		util.AssertNil(t,err)
		util.AssertEqual(t,string(result.Data),util.MakeInt(i,2*util.MAXCHUNKSIZE))
	}
	logrus.Infoln("score : 25")

	// 4 clients write to same file (independent chunk)
	wg.Add(4)
	data := make([]byte,4*util.MAXCHUNKSIZE)
	err := util.HTTPWrite(string(cList[0].GetClientAddr()),fdList[0][0],0, data)
	util.AssertNil(t,err)
	for i:=0;i<4;i++ {
		go func(index int) {
			clientAddr:= cList[index].GetClientAddr()
			offset := index*util.MAXCHUNKSIZE
			data = []byte(util.MakeInt(index,util.MAXCHUNKSIZE))
			err = util.HTTPWrite(string(clientAddr),fdList[index][0],offset,data)
			util.AssertNil(t,err)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i:=0;i<4;i++ {
		offset := i*util.MAXCHUNKSIZE
		result,err := util.HTTPRead(string(cList[0].GetClientAddr()),fdList[i][0],offset,util.MAXCHUNKSIZE)
		util.AssertNil(t,err)
		util.AssertEqual(t,string(result.Data),util.MakeInt(i,util.MAXCHUNKSIZE))
	}
	logrus.Infoln("score : 50")

	// 4 clients write to same file (cross chunk)
	wg.Add(4)
	data = make([]byte,6*util.MAXCHUNKSIZE)
	err = util.HTTPWrite(string(cList[0].GetClientAddr()),fdList[0][1],0, data)
	util.AssertNil(t,err)
	for i:=0;i<4;i++ {
		go func(index int) {
			offset := index*(util.MAXCHUNKSIZE*1.5)
			data = []byte(util.MakeInt(index,util.MAXCHUNKSIZE*1.5))
			err = util.HTTPWrite(string(cList[index].GetClientAddr()),fdList[index][1],offset,data)
			util.AssertNil(t,err)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i:=0;i<4;i++ {
		offset := i*(util.MAXCHUNKSIZE*1.5)
		result,err := util.HTTPRead(string(cList[0].GetClientAddr()),fdList[0][1],offset,util.MAXCHUNKSIZE*1.5)
		util.AssertNil(t,err)
		util.AssertEqual(t,string(result.Data),util.MakeInt(i,util.MAXCHUNKSIZE*1.5))
	}

	logrus.Infoln("score : 75")

	// 4 clients write to same file same place
	wg.Add(4)
	for i:=0;i<4;i++ {
		go func(index int) {
			clientAddr:= cList[index].GetClientAddr()
			offset := 0
			data = []byte(util.MakeInt(index,util.MAXCHUNKSIZE*4))
			err = util.HTTPWrite(string(clientAddr),fdList[index][1],offset,data)
			util.AssertNil(t,err)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i:=0;i<4;i++ {
		offset := i * util.MAXCHUNKSIZE
		result,err := util.HTTPRead(string(cList[0].GetClientAddr()),fdList[i][1],offset,util.MAXCHUNKSIZE)
		util.AssertNil(t,err)
		util.AssertSameData(t,result.Data)
	}
	logrus.Infoln("score : 100")
}

func CClear() {
	if true {
		for {
			err := os.RemoveAll("ck")
			if err != nil {
				logrus.Print(err)
			} else {
				logrus.Printf("Clear all\n")
				break
			}
			time.Sleep(time.Second)
		}
	}
}
func TestOpenClose(t *testing.T) {
	c,m,cs := InitTest()
	err := util.HTTPCreate(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	err = util.HTTPCreate(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)
	fd,err := util.HTTPOpen(util.CLIENTADDR,"/file1")
	util.AssertNil(t,err)
	logrus.Infoln("fd :",fd)
	code,err := util.HTTPClose(util.CLIENTADDR,fd)
	util.AssertEqual(t,code,200)
	fd,err = util.HTTPOpen(util.CLIENTADDR,"/file2")
	util.AssertNil(t,err)
	logrus.Infoln("fd :",fd)
	code,err = util.HTTPClose(util.CLIENTADDR,fd)
	util.AssertEqual(t,code,200)
	util.AssertNil(t,err)
	code,err = util.HTTPClose(util.CLIENTADDR,fd)
	util.AssertEqual(t,code,400)
	m.Exit()
	c.Exit()
	CClear()
	for _,_cs := range cs{
		_cs.Exit()
	}
}


