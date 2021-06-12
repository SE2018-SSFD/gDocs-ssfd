package main

import (
	"DFS/master"
	"DFS/util"
	"fmt"
	"testing"
)
func initMasterTest()(m *master.Master) {
	m = master.InitMaster("127.0.0.1:1234", ".")
	return m
}

//func TestNamespaceParallel(t *testing.T){
//	var wg sync.WaitGroup
//	m := initMasterTest()
//	wg.Add(util.PARALLELSIZE)
//	for i:=0 ;i<util.PARALLELSIZE;i++{
//		go _TestNamespaceParallel(t,m,&wg)
//	}
//	wg.Wait()
//}
//
//func _TestNamespaceParallel(t *testing.T,m *master.Master,wg *sync.WaitGroup){
//
//	wg.Done()
//}

func TestNamespaceSingle(t *testing.T) {
	m := initMasterTest()
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
	m.Exit()
}


