package test

import (
	"DFS/chunkserver"
	"DFS/client"
	"DFS/master"
	"DFS/util"
	"DFS/util/zkWrap"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

var HBisClear bool = true

func ClearCS() {
	if HBisClear {
		for {
			err := os.RemoveAll("cs")
			if err != nil {
				logrus.Print(err)
			} else {
				logrus.Printf("Clear cs\n")
				break
			}
		}
	}
}

func ClearLog() {
	if HBisClear {
		for {
			err := os.RemoveAll("log")
			if err != nil {
				logrus.Print(err)
			} else {
				logrus.Printf("Clear log\n")
				break
			}
		}
	}
}

// initTest init and return the handle of client,master and chunkservers
func InitHETest() (c []*client.Client, m []*master.Master, cs []*chunkserver.ChunkServer) {
	logrus.SetLevel(logrus.DebugLevel)
	// prepare log dir
	filename := "log/"
	_, err := os.Stat(filename)
	if err == nil {
		err := os.RemoveAll(filename)
		if err != nil {
			logrus.Fatalf("remove %v error\n", filename)
		}
	}
	err = os.Mkdir(filename, 0755)
	if err != nil {
		logrus.Fatalf("mkdir %v error\n", filename)
	}
	for i := 0; i < 3; i++ {
		filename := "log/log_" + strconv.Itoa(i+1)
		_, err := os.Stat(filename)
		if err == nil {
			err := os.RemoveAll(filename)
			if err != nil {
				logrus.Fatalf("remove %v error\n", filename)
			}
		}
		err = os.Mkdir(filename, 0755)
		if err != nil {
			logrus.Fatalf("mkdir %v error\n", filename)
		}
	}
	// Init three masters
	m = make([]*master.Master, 3)
	addrList := make([]string, 3)
	addrList[0] = util.MASTER1ADDR
	addrList[1] = util.MASTER2ADDR
	addrList[2] = util.MASTER3ADDR
	var wg sync.WaitGroup
	wg.Add(util.MASTERCOUNT)

	// make the first master be a leader
	var mp *master.Master
	mp, err = master.InitMultiMaster(util.Address(addrList[0]), util.LinuxPath("log/log_"+strconv.Itoa(1)))
	if err != nil {
		fmt.Println("Master Init Error :", err)
		os.Exit(1)
	}
	m[0] = mp
	wg.Done()
	go func() {
		m[0].Serve()
	}()

	// start other two masters
	for i := 1; i < util.MASTERCOUNT; i++ {
		go func(order int) {
			var mp *master.Master
			mp, err := master.InitMultiMaster(util.Address(addrList[order]), util.LinuxPath("log/log_"+strconv.Itoa(order+1)))
			if err != nil {
				fmt.Println("Master Init Error :", err)
				os.Exit(1)
			}
			m[order] = mp
			wg.Done()
			mp.Serve()
		}(i)
	}
	wg.Wait()
	//Init five clients
	c = make([]*client.Client, 5)
	for index, port := range []int{1333, 1334, 1335, 1336, 1337} {
		go func(order int, p int) {
			addr := util.Address("127.0.0.1:" + strconv.Itoa(p))
			c[order] = client.InitClient(addr, util.Address(""))
			c[order].Serve()
		}(index, port)
	}

	// Register some virtual chunkServers
	_, err = os.Stat("cs")
	if err == nil {
		err := os.RemoveAll("cs")
		if err != nil {
			logrus.Fatalf("mkdir %v error\n", "cs")
		}
	}
	err = os.Mkdir("cs", 0755)
	if err != nil {
		logrus.Fatalf("mkdir %v error\n", "cs")
	}
	cs = make([]*chunkserver.ChunkServer, 5)
	for index, port := range []int{3000, 3001, 3002, 3003, 3004} {
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		cs[index] = chunkserver.InitChunkServer(string(addr), "cs/cs"+strconv.Itoa(port), util.MASTER1ADDR)
	}

	time.Sleep(500 * time.Millisecond)

	// Test whether master addrs are the same

	return
}

func TestElectionAndHeartbeat(t *testing.T) {
	zkWrap.Chroot("/DFS")
	c, m, _ := InitHETest()
	// should print "Master Addr : 127.0.0.1:1234"
	for i := 0; i < 5; i++ {
		c[i].PrintMasterAddr()
	}

	// now master0 crash,leader should change
	m[0].Exit()
	// m[0].
	time.Sleep(10 * time.Second)
	for i := 0; i < 5; i++ {
		c[i].PrintMasterAddr()
	}
	ClearCS()
	ClearLog()
}
