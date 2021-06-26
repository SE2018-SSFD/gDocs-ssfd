package main

import (
	"DFS/chunkserver"
	"DFS/util"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

var isClear bool = false //whether clear all file after test

func initChunkServer() (cs []*chunkserver.ChunkServer) {
	logrus.SetLevel(logrus.DebugLevel)
	// Register some virtual chunkServers
	_, err := os.Stat("cs")
	if err != nil {
		err := os.Mkdir("cs", 0755)
		if err != nil {
			logrus.Fatalf("mkdir %v error\n", "cs")
		}
	}
	cs = make([]*chunkserver.ChunkServer, 5)
	for index, port := range []int{3000, 3001, 3002, 3003, 3004} {
		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
		cs[index] = chunkserver.InitChunkServer(string(addr), "cs/cs"+strconv.Itoa(port), util.MASTERADDR)
	}

	time.Sleep(time.Second)
	return
}

func chunkServerExit(cs []*chunkserver.ChunkServer) {
	for _, c := range cs {
		c.Exit()
	}
}
func chunkServerCrash(cs []*chunkserver.ChunkServer) {
	for _, c := range cs {
		c.Crash()
	}
}

func Clear() {

	if isClear {
		for {
			err := os.RemoveAll("cs")
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

func TestChunkServer(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	cs := initChunkServer()

	var h []util.Handle = make([]util.Handle, 2)
	var wg []sync.WaitGroup = make([]sync.WaitGroup, 2)
	var createChunkArgs []util.CreateChunkArgs = make([]util.CreateChunkArgs, 2)
	h[0] = util.Handle(1)
	h[1] = util.Handle(2)
	wg[0].Add(3)
	wg[1].Add(3)
	createChunkArgs[0] = util.CreateChunkArgs{Handle: h[0]}
	createChunkArgs[1] = util.CreateChunkArgs{Handle: h[1]}
	for i := 0; i < 3; i++ {
		go func(idx int) {
			err := cs[idx].CreateChunkRPC(createChunkArgs[0], nil)
			if err != nil {
				logrus.Printf("CreateChunk error\n")
				logrus.Println(err)
				t.Fail()
			}
			wg[0].Done()
		}(i)
	}
	for i := 1; i < 4; i++ {
		go func(idx int) {
			err := cs[idx].CreateChunkRPC(createChunkArgs[1], nil)
			if err != nil {
				logrus.Printf("CreateChunk error\n")
				logrus.Println(err)
				t.Fail()
			}
			wg[1].Done()
		}(i)
	}
	wg[0].Wait()
	wg[1].Wait()
	logrus.Printf("Create Chunk Done\n")

	var str string = "abcdefg"

	data := []byte(str)
	var csAddrs [][]util.Address = make([][]util.Address, 2)
	csAddrs[0] = make([]util.Address, 0)
	csAddrs[1] = make([]util.Address, 0)

	//notice i begin with 1
	for i := 1; i < 3; i++ {
		csAddrs[0] = append(csAddrs[0], util.Address(cs[i].GetAddr()))
	}
	for i := 2; i < 4; i++ {
		csAddrs[1] = append(csAddrs[1], util.Address(cs[i].GetAddr()))
	}
	var cid []util.CacheID = make([]util.CacheID, 4)
	cid[0] = util.CacheID{Handle: h[0], ClientAddr: "127.0.0.1:2100"}
	cid[1] = util.CacheID{Handle: h[0], ClientAddr: "127.0.0.1:2101"}
	cid[2] = util.CacheID{Handle: h[1], ClientAddr: "127.0.0.1:2102"}
	cid[3] = util.CacheID{Handle: h[1], ClientAddr: "127.0.0.1:2103"}
	wg[0].Add(4)
	// load Data
	var loadArgs []util.LoadDataArgs = make([]util.LoadDataArgs, 4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			loadArgs[idx] = util.LoadDataArgs{
				Data:  data,
				CID:   cid[idx],
				Addrs: csAddrs[idx/2]}
			err := cs[idx/2].LoadDataRPC(loadArgs[idx], nil)
			if err != nil {
				logrus.Printf("loadData error\n")
				logrus.Println(err)
				t.Fail()
			}
			wg[0].Done()
		}(i)
	}
	wg[0].Wait()
	logrus.Println("Load Data Done!")

	wg[0].Add(4)
	var syncArgs []util.SyncArgs = make([]util.SyncArgs, 4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			syncArgs[idx] = util.SyncArgs{CID: cid[idx], Off: (idx % 2) * 4, Addrs: csAddrs[idx/2]}
			err := cs[idx/2].SyncRPC(syncArgs[idx], nil)
			if err != nil {
				logrus.Printf("Sync error\n")
				logrus.Println(err)
				t.Fail()
			}
			wg[0].Done()
		}(i)
	}
	wg[0].Wait()
	logrus.Printf("Write Done !!!\n")

	var readArgs []util.ReadChunkArgs = make([]util.ReadChunkArgs, 4)
	var readReply []util.ReadChunkReply = make([]util.ReadChunkReply, 4)
	wg[0].Add(4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			readArgs[idx] = util.ReadChunkArgs{Handle: h[idx/2], Off: idx % 2, Len: 6}
			err := cs[idx/2+1].ReadChunkRPC(readArgs[idx], &readReply[idx])
			if err != nil {
				logrus.Printf("Read error\n")
				logrus.Println(err)
				t.Fail()
			} else {
				logrus.Printf("read success, idx is %v, data is %s\n", idx, string(readReply[idx].Buf[:]))
			}
			wg[0].Done()
		}(i)
	}
	wg[0].Wait()
	logrus.Printf("Read Done !!!\n")

	// test log and read
	chunkServerCrash(cs)
	cs = initChunkServer()
	wg[0].Add(4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			readArgs[idx] = util.ReadChunkArgs{Handle: h[idx/2], Off: idx % 2, Len: 6}
			err := cs[idx/2+1].ReadChunkRPC(readArgs[idx], &readReply[idx])
			if err != nil {
				logrus.Printf("Read error\n")
				logrus.Println(err)
				t.Fail()
			} else {
				logrus.Printf("read success, idx is %v, data is %s\n", idx, string(readReply[idx].Buf[:]))
			}
			wg[0].Done()
		}(i)
	}
	wg[0].Wait()
	logrus.Printf("Recover and Read Done !!!\n")

	// test checkpoint and read
	chunkServerExit(cs)
	cs = initChunkServer()
	wg[0].Add(4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			readArgs[idx] = util.ReadChunkArgs{Handle: h[idx/2], Off: idx % 2, Len: 6}
			err := cs[idx/2+1].ReadChunkRPC(readArgs[idx], &readReply[idx])
			if err != nil {
				logrus.Printf("Read error\n")
				logrus.Println(err)
				t.Fail()
			} else {
				logrus.Printf("read success, idx is %v, data is %s\n", idx, string(readReply[idx].Buf[:]))
			}
			wg[0].Done()
		}(i)
	}
	wg[0].Wait()
	logrus.Printf("Recover and Read Done !!!\n")

	chunkServerExit(cs)
	Clear()
}
