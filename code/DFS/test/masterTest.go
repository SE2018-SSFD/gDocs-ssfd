package main

import (
	"DFS/master"
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
)


// TODO : use go testing package to rewrite an assert-style program
func main() {
	logrus.SetLevel(logrus.DebugLevel)
	m := master.InitMaster("127.0.0.1:1234", ".")
	NamespaceTest(m)
}

func NamespaceTest(m *master.Master) {
	var createReply util.CreateRet
	var mkdirReply util.MkdirRet
	var listReply util.ListRet
	var getFileMetaReply util.GetFileMetaRet
	err := m.CreateRPC(util.CreateArg{Path: "/file1"}, &createReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.CreateRPC(util.CreateArg{Path: "/file1"}, &createReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.MkdirRPC(util.MkdirArg{Path: "/dir1"}, &mkdirReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.MkdirRPC(util.MkdirArg{Path: "/dir1"}, &mkdirReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file1"}, &createReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.CreateRPC(util.CreateArg{Path: "/dir1/file2"}, &createReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.CreateRPC(util.CreateArg{Path: "/nonexist/file1"}, &createReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.CreateRPC(util.CreateArg{Path: "/invalidPath/"}, &createReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.CreateRPC(util.CreateArg{Path: "invalidPath/file1"}, &createReply)
	if err != nil {
		fmt.Println(err)
	}
	err = m.ListRPC(util.ListArg{Path: "/dir1"}, &listReply)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, file := range listReply.Files {
			fmt.Print(file, " ")
		}
		fmt.Println()
	}
	err = m.GetFileMetaRPC(util.GetFileMetaArg{Path: "/dir1"}, &getFileMetaReply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getFileMetaReply.Exist, " ", getFileMetaReply.IsDir, " ", getFileMetaReply.Size)
	}
	err = m.GetFileMetaRPC(util.GetFileMetaArg{Path: "/dir1/file1"}, &getFileMetaReply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getFileMetaReply.Exist, " ", getFileMetaReply.IsDir, " ", getFileMetaReply.Size)
	}
}
