package main

import (
	"DFS/chunkserver"
	"DFS/client"
	"DFS/master"
	"DFS/util"
	"DFS/util/zkWrap"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}
	setLoggingStrategy()
	switch os.Args[1] {
	case "master":
		singleMasterRun()
	case "multimaster":
		multiMasterRun()
	case "chunkServer":
		chunkServerRun()
	case "client":
		clientRun()
	default:
		fmt.Println("Unsupported node type!")
		printUsage()
		return
	}

}

func clientRun() {
	if len(os.Args) < 4 {
		printUsage()
		return
	}
	addr := os.Args[2]
	masterAddr := os.Args[3]
	c := client.InitClient(util.Address(addr), util.Address(masterAddr))
	c.Serve()
}

func chunkServerRun() {
	if len(os.Args) < 5 {
		printUsage()
		return
	}
	addr := os.Args[2]
	dataPath := os.Args[3]
	masterAddr := os.Args[4]
	_ = chunkserver.InitChunkServer(addr, dataPath, masterAddr)
	// block on ch; make it a daemon
	ch := make(chan bool)
	<-ch
}
func singleMasterRun(){
	if len(os.Args) < 4 {
		printUsage()
		return
	}
	addr := os.Args[2]
	dataPath := os.Args[3]
	m,_ := master.InitMaster(util.Address(addr), util.LinuxPath(dataPath))
	m.Serve()
	// block on ch; make it a daemon
	ch := make(chan bool)
	<-ch
}

func multiMasterRun() {
	if len(os.Args) < 8 {
		printUsage()
		return
	}
	err := zkWrap.Chroot("/DFS")
	if err != nil {
		return 
	}
	//arg 2,3,4 are addresses of master;arg 5,6,7 are metadata paths of master
	for i:=0;i<util.MASTERCOUNT;i++{
		go func(order int) {
			m,err := master.InitMultiMaster(util.Address(os.Args[2+order]), util.LinuxPath(os.Args[5+order]))
			if err!=nil{
				fmt.Println("Master Init Error :",err)
				os.Exit(1)
			}
			m.Serve()
		}(i)
	}
	ch := make(chan bool)
	<-ch
}

// set the default logging strategy of DFS
func setLoggingStrategy() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" NodeRunner master <addr1> <metapath1>")
	fmt.Println(" NodeRunner multimaster <addr1> <addr2> <addr3> <metapath1> <metapath2> <metapath3>")
	fmt.Println(" NodeRunner chunkServer <addr> <root path> <master addr>")
	fmt.Println(" NodeRunner client <addr> <master addr>")
	fmt.Println()
}
