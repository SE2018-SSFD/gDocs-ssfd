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
	"testing"
	"time"
)

// initTest init and return the handle of client,master and chunkservers
func InitTest() (c *client.Client,m *master.Master,cs []*chunkserver.ChunkServer){
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
	if err != nil {
		err := os.Mkdir("ck", 0755)
		if err != nil {
			logrus.Fatalf("mkdir %v error\n", "ck")
		}
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

func TestReadWrite(t *testing.T) {
	c,m,cs := InitTest()
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

	m.Exit()
	c.Exit()
	for _,_cs := range cs{
		_cs.Exit()
	}

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


