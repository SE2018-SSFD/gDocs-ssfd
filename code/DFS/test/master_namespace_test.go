package main

import (
	"DFS/master"
	"DFS/util"
	"fmt"
	"testing"
)


func TestNamespace(t *testing.T) {
	m := master.InitMaster("127.0.0.1:1234", ".")
	var createReply util.CreateRet
	var mkdirReply util.MkdirRet
	var listReply util.ListRet
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
}


