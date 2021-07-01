package chunkserver

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sync"
)

type ChunkServer struct {
	addr             string
	masterAddr       string
	dir              string
	l                net.Listener
	sync.RWMutex                // protect chunks
	logLock          sync.Mutex // protect log
	chunks           map[util.Handle]*ChunkInfo
	cache            *Cache
	shutdown         chan struct{}
	clusterHeartbeat *zkWrap.Heartbeat
}

type ChunkInfo struct {
	sync.RWMutex
	verNum    util.Version //version number
	mutations map[int64]*Mutation
	isStale   bool
	// checksum  int64
}

type OperationType int32

const (
	Operation_Delete OperationType = 0
	Operation_Update OperationType = 1
)

type ChunkInfoCP struct {
	Handle util.Handle
	VerNum util.Version
}

type ChunkInfoLog struct {
	Handle    util.Handle
	VerNum    util.Version
	Operation OperationType
}

type Mutation struct {
	Checksum int32
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

	cs.RecoverChunkInfo()
	log.Printf("chunkserver %v: init success\n", chunkAddr)

	cs.StartRPCServer()

	return cs
}

func (cs *ChunkServer) Exit() {
	err := cs.l.Close()
	close(cs.shutdown)
	cs.StoreCheckPoint()
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
