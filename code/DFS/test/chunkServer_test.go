package main

import (
	"DFS/chunkserver"
	"DFS/util"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestChunkServer(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	// master.InitMaster("127.0.0.1:1234", ".")
	// client.InitClient("127.0.0.1:2100")
	cs := make([]*chunkserver.ChunkServer, 3)
	cs[0] = chunkserver.InitChunkServer("127.0.0.1:2000", "ck2000", "127.0.0.1:1234")
	cs[1] = chunkserver.InitChunkServer("127.0.0.1:2001", "ck2001", "127.0.0.1:1234")
	cs[2] = chunkserver.InitChunkServer("127.0.0.1:2002", "ck2002", "127.0.0.1:1234")

	var h = util.Handle(1)
	var wg sync.WaitGroup
	wg.Add(3)

	createArgs := util.CreateChunkArgs{Handle: h}
	for i := 0; i < 3; i++ {
		go func(idx int) {
			err := cs[idx].CreateChunkRPC(createArgs, nil)
			if err != nil {
				fmt.Printf("CreateChunk error\n")
				fmt.Println(err)
				t.Fail()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("Create Chunk Done\n")

	var str string = "test"

	data := []byte(str)
	csAddrs := make([]util.Address, 0)

	//notic i begin with 1
	for i := 1; i < 3; i++ {
		csAddrs = append(csAddrs, util.Address(cs[i].GetAddr()))
	}
	cid := util.CacheID{Handle: h, ClientAddr: "127.0.0.1:2100"}
	var loadArgs = util.LoadDataArgs{
		Data:  data,
		CID:   cid,
		Addrs: csAddrs}
	err := cs[0].LoadDataRPC(loadArgs, nil)
	if err != nil {
		fmt.Printf("loadData error\n")
		fmt.Println(err)
		t.Fail()
	}

	var syncArgs = util.SyncArgs{CID: cid, Off: 0, Addrs: csAddrs}
	err = cs[0].SyncRPC(syncArgs, nil)
	if err != nil {
		fmt.Printf("Sync error\n")
		fmt.Println(err)
		t.Fail()
	}
	fmt.Printf("Write Done !!!\n")
	var readArgs = util.ReadChunkArgs{Handle: h, Off: 0, Len: 4}
	var readReply util.ReadChunkReply
	err = cs[0].ReadChunkRPC(readArgs, &readReply)
	if err != nil {
		fmt.Printf("Read error\n")
		fmt.Println(err)
		t.Fail()
	} else {
		fmt.Printf("read success, data is %s\n", string(readReply.Buf[:]))
	}
	Clear(cs)
}

func Clear(cs []*chunkserver.ChunkServer) {
	for i := 0; i < 3; i++ {
		cs[i].RemoveChunk(util.Handle(1))
	}
	os.Remove("ck2000")
	os.Remove("ck2001")
	os.Remove("ck2002")

	fmt.Printf("Clear all\n")
}
