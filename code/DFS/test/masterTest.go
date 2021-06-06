package main

import (
	"DFS/master"
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
)

func main(){
	//str := "/a/b/"
	//for index,symbol := range strings.FieldsFunc(str,func(c rune) bool {return c=='/'}){
	//	fmt.Println(index," ",symbol)
	//}
	logrus.SetLevel(logrus.DebugLevel)
	m := master.InitMaster("127.0.0.1:1234", ".")
	NamespaceTest(m)
}

func NamespaceTest(m *master.Master){
	var reply util.CreateRet
	err := m.CreateRPC(util.CreateArg{Path: "/abc"},&reply)
	fmt.Println(err)
	err = m.CreateRPC(util.CreateArg{Path: "/abc"},&reply)
	fmt.Println(err)
}