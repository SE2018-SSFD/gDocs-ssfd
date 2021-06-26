package main

import (
	"DFS/master"
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)
func initMasterTest()(m *master.Master) {
	m,err := master.InitMaster("127.0.0.1:1234", ".")
	if err!=nil{
		os.Exit(1)
	}
	return m
}

func testMkdir(t *testing.T,path util.DFSPath,m *master.Master,ignore bool,count *int,lock *sync.Mutex){
	var mkdirReply util.MkdirRet
	err := m.MkdirRPC(util.MkdirArg{Path: path},&mkdirReply)
	if err != nil && ignore{
		lock.Lock()
		*count += 1
		lock.Unlock()
		return
	}else{
		util.AssertNil(t,err)
	}
	return
}

func testDelete(t *testing.T,path util.DFSPath,m *master.Master,ignore bool,count *int,lock *sync.Mutex){
	var deleteReply util.DeleteRet
	err := m.DeleteRPC(util.DeleteArg{Path: path},&deleteReply)
	if err != nil && ignore{
		lock.Lock()
		*count += 1
		lock.Unlock()
		return
	}else{
		util.AssertNil(t,err)
	}
	return
}

func testCreate(t *testing.T,path util.DFSPath,m *master.Master,ignore bool,count *int,lock *sync.Mutex){
	var createReply util.CreateRet
	err := m.CreateRPC(util.CreateArg{Path: path},&createReply)
	if err != nil && ignore{
		lock.Lock()
		*count += 1
		lock.Unlock()
		return
	}else{
		util.AssertNil(t,err)
	}
	return
}

func TestNamespaceParallel(t *testing.T){
	count := 0
	score := 0
	var wg sync.WaitGroup // To sync all the goroutines
	lock := sync.Mutex{} // To protect the count
	var mkdirReply util.MkdirRet
	m := initMasterTest()
	defer m.Exit()
	// Test concurrent create and delete on different file
	wg.Add(util.PARALLELSIZE)
	for i:=0 ;i<util.PARALLELSIZE;i++{
		go func(i int){
			path := util.DFSPath("/file"+strconv.Itoa(i))
			testCreate(t,path,m,false,&count,&lock)
			testDelete(t,path,m,false,&count,&lock)
			testCreate(t,path,m,false,&count,&lock)
			testDelete(t,path,m,false,&count,&lock)
			wg.Done()
		}(i)
	}
	implicitWait(util.MAXWAITINGTIME * time.Second,&wg)
	util.AssertEqual(t,count,0)
	logrus.Infof("Score %d\n",score)
	score += 20

	// Test concurrent create and delete on deep different file
	err := m.MkdirRPC(util.MkdirArg{Path: "/dir1"}, &mkdirReply)
	util.AssertNil(t,err)
	err = m.MkdirRPC(util.MkdirArg{Path: "/dir1/secondDir1"}, &mkdirReply)
	util.AssertNil(t,err)
	wg.Add(util.PARALLELSIZE)
	for i:=0 ;i<util.PARALLELSIZE;i++{
		go func(i int){
			path := util.DFSPath("/dir1/secondDir1/file"+strconv.Itoa(i))
			testCreate(t,path,m,false,&count,&lock)
			testDelete(t,path,m,false,&count,&lock)
			testCreate(t,path,m,false,&count,&lock)
			testDelete(t,path,m,false,&count,&lock)
			wg.Done()
		}(i)
	}
	implicitWait(util.MAXWAITINGTIME * time.Second,&wg)
	util.AssertEqual(t,count,0)
	score += 20
	logrus.Infof("Score %d\n",score)

	// Test concurrent create and delete on deep same file
	wg.Add(util.PARALLELSIZE)
	for i:=0 ;i<util.PARALLELSIZE;i++{
		go func(i int){
			path := util.DFSPath("/dir1/secondDir1/filesame")
			testCreate(t,path,m,true,&count,&lock)
			wg.Done()
		}(i)
	}
	implicitWait(util.MAXWAITINGTIME * time.Second,&wg)
	wg.Add(util.PARALLELSIZE)
	for i:=0 ;i<util.PARALLELSIZE;i++{
		go func(i int){
			path := util.DFSPath("/dir1/secondDir1/filesame")
			testDelete(t,path,m,true,&count,&lock)
			wg.Done()
		}(i)
	}
	implicitWait(util.MAXWAITINGTIME * time.Second,&wg)
	util.AssertEqual(t,count,2*(util.PARALLELSIZE-1))
	score += 20
	logrus.Infof("Score %d\n",score)

	// Test concurrent create and delete on random different files
	count = 0
	wg.Add(util.PARALLELSIZE)
	for i:=0 ;i<util.PARALLELSIZE;i++{
		go func(i int){
			path := util.DFSPath("/dir1/secondDir1/"+util.MakeString(rand.Int()%20+1)+strconv.Itoa(i))
			testCreate(t,path,m,true,&count,&lock)
			path2 := util.DFSPath("/dir1/"+util.MakeString(rand.Int()%20+1)+strconv.Itoa(i))
			testCreate(t,path2,m,true,&count,&lock)
			dir :=  util.DFSPath("/dir1/"+strconv.Itoa(i))
			testMkdir(t,dir,m,true,&count,&lock)
			path3 := util.DFSPath(string(dir)+util.MakeString(rand.Int()%20+1)+strconv.Itoa(i))
			testCreate(t,path3,m,true,&count,&lock)
			testDelete(t,path,m,true,&count,&lock)
			testDelete(t,path2,m,true,&count,&lock)
			testDelete(t,path3,m,true,&count,&lock)
			wg.Done()
		}(i)
	}
	implicitWait(util.MAXWAITINGTIME * time.Second,&wg)
	util.AssertEqual(t,count,0)
	score += 40
	logrus.Infof("Score %d\n",score)
}

func implicitWait(t time.Duration,wg *sync.WaitGroup){
	c := make(chan int)
	go func(){
		wg.Wait()
		c <- 0
	}()
	select {
	case  <-c:
	case <-time.After(t):
		logrus.Fatalf("Parallel test failed\n")
	}
}


func TestNamespaceSingle(t *testing.T) {
	m := initMasterTest()
	defer m.Exit()
	var createReply util.CreateRet
	var mkdirReply util.MkdirRet
	var listReply util.ListRet
	var deleteReply util.DeleteRet
	var getFileMetaReply util.GetFileMetaRet
	err := m.CreateRPC(util.CreateArg{Path: "/file1"}, &createReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/file1"}, &createReply)
	util.AssertNotNil(t,err)
	err = m.MkdirRPC(util.MkdirArg{Path: "/dir1"}, &mkdirReply)
	util.AssertNil(t,err)
	err = m.MkdirRPC(util.MkdirArg{Path: "/dir1"}, &mkdirReply)
	util.AssertNotNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file1"}, &createReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file2"}, &createReply)
	util.AssertNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/nonexist/file1"}, &createReply)
	util.AssertNotNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "/invalidPath/"}, &createReply)
	util.AssertNotNil(t,err)
	err = m.CreateRPC(util.CreateArg{Path: "invalidPath/file1"}, &createReply)
	util.AssertNotNil(t,err)
	err = m.ListRPC(util.ListArg{Path: "/dir1"}, &listReply)
	util.AssertNil(t,err)
	for _, file := range listReply.Files {
		fmt.Print(file, " ")
	}
	fmt.Println()
	err = m.GetFileMetaRPC(util.GetFileMetaArg{Path: "/dir1"}, &getFileMetaReply)
	util.AssertNil(t,err)
	fmt.Println(getFileMetaReply.Exist, " ", getFileMetaReply.IsDir, " ", getFileMetaReply.Size)
	err = m.GetFileMetaRPC(util.GetFileMetaArg{Path: "/dir1/file1"}, &getFileMetaReply)
	util.AssertNil(t,err)
	fmt.Println(getFileMetaReply.Exist, " ", getFileMetaReply.IsDir, " ", getFileMetaReply.Size)
	err = m.DeleteRPC(util.DeleteArg{Path: "/dir1/file1"},&deleteReply)
	err = m.GetFileMetaRPC(util.GetFileMetaArg{Path: "/dir1/file1"}, &getFileMetaReply)
	fmt.Println(getFileMetaReply.Exist, " ", getFileMetaReply.IsDir, " ", getFileMetaReply.Size)

	util.AssertEqual(t,getFileMetaReply.Exist,false)
}


