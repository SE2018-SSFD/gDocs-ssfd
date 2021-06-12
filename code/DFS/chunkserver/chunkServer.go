package chunkserver

import (
	"DFS/util"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sync"
)

type ChunkServer struct {
	addr       string
	masterAddr string
	dir        string
	l          net.Listener
	sync.RWMutex

	chunks   map[util.Handle]*ChunkInfo
	cache    *Cache
	shutdown chan struct{}
}

type ChunkInfo struct {
	sync.RWMutex
	verNum    int64 //version number
	mutations map[int64]*Mutation
	// checksum  int64
}

type ChunkInfoCP struct {
	handle util.Handle
	verNum int64
}

type Mutation struct {
}

func InitChunkServer(chunkAddr string, dataPath string, masterAddr string) *ChunkServer {
	cs := &ChunkServer{
		addr:       chunkAddr,
		dir:        dataPath,
		masterAddr: masterAddr,
		cache:      InitCache(),
		chunks:     make(map[util.Handle]*ChunkInfo),
		shutdown:   make(chan struct{}),
	}

	_, err := os.Stat(cs.dir)
	if err != nil {
		err := os.Mkdir(cs.dir, 0755)
		if err != nil {
			log.Fatalf("mkdir %v error\n", cs.dir)
		}
	}

	cs.StartRPCServer()

	log.Printf("chunkserver %v: init success\n", chunkAddr)
	return cs
}

//TODO: should set checkpoint
func (cs *ChunkServer) Exit() {
	err := cs.l.Close()
	close(cs.shutdown)
	if err != nil {
		return
	}
}

func (cs *ChunkServer) GetAddr() string {
	return string(cs.addr)
}

func (cs *ChunkServer) GetFileName(handle util.Handle) string {
	name := fmt.Sprintf("chunk-%v.dat", handle)
	return path.Join(cs.dir, name)
}

func (cs *ChunkServer) GetCPName() string {
	return path.Join(cs.dir, "checkpoint.dat")
}

func (cs *ChunkServer) GetLogName() string {
	return path.Join(cs.dir, "log.dat")
}
