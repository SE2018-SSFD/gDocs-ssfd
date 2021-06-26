package main

import (
	"DFS/chunkserver"
	"DFS/client"
	"DFS/master"
	"DFS/util"
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
		masterRun()
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
	c := chunkserver.InitChunkServer(addr, dataPath, masterAddr)
	// logrus.Info(c.GetStatusString())
	// block on ch; make it a daemon
	ch := make(chan bool)
	<-ch
}

func masterRun() {
	if len(os.Args) < 4 {
		printUsage()
		return
	}
	addr := util.Address(os.Args[2])
	metaPath := util.LinuxPath(os.Args[3])
	m := master.InitMaster(addr, metaPath)
	logrus.Info(m.GetStatusString())
	m.Serve()
}

// set the default logging strategy of DFS
func setLoggingStrategy() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" NodeRunner master <addr> <root path>")
	fmt.Println(" NodeRunner chunkServer <addr> <root path> <master addr>")
	fmt.Println(" NodeRunner client <addr> <master addr>")
	fmt.Println()
}
