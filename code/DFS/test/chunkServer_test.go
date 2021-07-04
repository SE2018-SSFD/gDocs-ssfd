package test

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

var isClear bool = true //whether clear all file after test

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
		cs[index] = chunkserver.InitChunkServer(string(addr), "cs/cs"+strconv.Itoa(port), util.MASTER1ADDR)
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

func TestReReplicate(t *testing.T) {
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
	//server 0,1,2 store chunk-1
	//server 1,2,3 store chunk-2
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
			syncArgs[idx] = util.SyncArgs{CID: cid[idx], Off: (idx % 2) * 4, Addrs: csAddrs[idx/2], IsAppend: false}
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

	// test clone chunk
	// clone chunk-1 from server 2 to server 3 & 4
	// clone chunk-2 from server 2 to server 4
	var cloneArgs []util.CloneChunkArgs = make([]util.CloneChunkArgs, 2)
	var cloneReply []util.CloneChunkReply = make([]util.CloneChunkReply, 2)
	var cloneAddr [][]util.Address = make([][]util.Address, 2)
	cloneAddr[0] = append(cloneAddr[0], util.Address(cs[4].GetAddr()))
	cloneAddr[0] = append(cloneAddr[0], util.Address(cs[3].GetAddr()))
	cloneAddr[1] = append(cloneAddr[1], util.Address(cs[4].GetAddr()))

	wg[0].Add(2)
	for i := 0; i < 2; i++ {
		go func(idx int) {
			cloneArgs[idx] = util.CloneChunkArgs{Handle: h[idx], Len: 4 + len(str), Addrs: cloneAddr[idx]}
			err := cs[2].CloneChunkRPC(cloneArgs[idx], &cloneReply[idx])
			if err != nil {
				logrus.Printf("Clone error\n")
				logrus.Println(err)
				t.Fail()
			}
			wg[0].Done()
		}(i)
	}
	wg[0].Wait()

	wg[0].Add(4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			readArgs[idx] = util.ReadChunkArgs{Handle: h[idx/2], Off: idx % 2, Len: 6}
			err := cs[idx/2+3].ReadChunkRPC(readArgs[idx], &readReply[idx])
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
	logrus.Printf("Clone and Read Done !!!\n")

	chunkServerExit(cs)
	Clear()
}

func TestConcurrentAppend(t *testing.T) {
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

	var str [][]byte = make([][]byte, 8)
	for i := 0; i < 8; i++ {
		str[i] = []byte(util.MakeInt(i, util.MAXCHUNKSIZE/4))
	}

	var csAddrs [][]util.Address = make([][]util.Address, 2)
	csAddrs[0] = make([]util.Address, 0)
	csAddrs[1] = make([]util.Address, 0)

	//notice i begin with 1
	//server 0,1,2 store chunk-1
	//server 1,2,3 store chunk-2
	for i := 1; i < 3; i++ {
		csAddrs[0] = append(csAddrs[0], util.Address(cs[i].GetAddr()))
	}
	for i := 2; i < 4; i++ {
		csAddrs[1] = append(csAddrs[1], util.Address(cs[i].GetAddr()))
	}
	var cid []util.CacheID = make([]util.CacheID, 8)
	cid[0] = util.CacheID{Handle: h[0], ClientAddr: "127.0.0.1:2100"}
	cid[1] = util.CacheID{Handle: h[0], ClientAddr: "127.0.0.1:2101"}
	cid[2] = util.CacheID{Handle: h[0], ClientAddr: "127.0.0.1:2102"}
	cid[3] = util.CacheID{Handle: h[0], ClientAddr: "127.0.0.1:2103"}
	cid[4] = util.CacheID{Handle: h[1], ClientAddr: "127.0.0.1:2100"}
	cid[5] = util.CacheID{Handle: h[1], ClientAddr: "127.0.0.1:2101"}
	cid[6] = util.CacheID{Handle: h[1], ClientAddr: "127.0.0.1:2102"}
	cid[7] = util.CacheID{Handle: h[1], ClientAddr: "127.0.0.1:2103"}
	// wg[0].Add(8)
	// load Data
	var loadArgs []util.LoadDataArgs = make([]util.LoadDataArgs, 8)
	for i := 0; i < 8; i++ {
		wg[0].Add(1)
		go func(idx int) {
			loadArgs[idx] = util.LoadDataArgs{
				Data:  str[idx],
				CID:   cid[idx],
				Addrs: csAddrs[idx/4]}
			err := cs[idx/4].LoadDataRPC(loadArgs[idx], nil)
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

	// wg[0].Add(4)
	var syncArgs []util.SyncArgs = make([]util.SyncArgs, 8)
	var syncReplys []util.SyncReply = make([]util.SyncReply, 8)
	for i := 0; i < 8; i++ {
		wg[0].Add(1)
		go func(idx int) {
			syncArgs[idx] = util.SyncArgs{CID: cid[idx], Off: 0, Addrs: csAddrs[idx/4], IsAppend: true}
			err := cs[idx/4].SyncRPC(syncArgs[idx], &syncReplys[idx])
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

	err := cs[0].LoadDataRPC(loadArgs[0], nil)
	if err != nil {
		logrus.Printf("loadData error\n")
		logrus.Println(err)
		t.Fail()
	}
	err = cs[0].SyncRPC(syncArgs[0], &syncReplys[0])
	if syncReplys[0].ErrorCode != util.NOSPACE {
		if err != nil {
			logrus.Printf("Sync error\n")
			logrus.Println(err)
		}
		t.Fail()
	}
	chunkServerExit(cs)
	Clear()
}

// func TestConcurrentAppend(t *testing.T) {
// 	filename := "test.txt"
// 	data := "123456"
// 	buf := []byte(data)
// 	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
// 	fd.WriteAt(buf, int64(5))
// 	if err != nil {
// 		logrus.Print(err)
// 	}
// 	var buf2 []byte = make([]byte, 10)
// 	n, err := fd.ReadAt(buf2, 0)
// 	logrus.Print(n)
// 	if err != nil {
// 		logrus.Print(err)
// 	}
// 	logrus.Print(buf2)
// }
