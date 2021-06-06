package chunkserver

import (
	"DFS/util"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path"
	"sync"
)

type ChunkServer struct {
	addr       string
	masterAddr string
	dir        string
	dataPath   string
	l          net.Listener
	lock       sync.RWMutex
}

func InitChunkServer(chunkAddr string, dataPath string, masterAddr string) *ChunkServer {
	cs := &ChunkServer{
		addr:       chunkAddr,
		dataPath:   dataPath,
		masterAddr: masterAddr,
	}
	rpcs := rpc.NewServer()
	rpcs.Register(cs)

	return cs
}
func (cs *ChunkServer) GetStatusString() string {
	return "ChunkServer address :" + cs.addr + ",dataPath :" + cs.dataPath
}

func (cs *ChunkServer) GetFileName(cid int64) string {
	name := fmt.Sprintf("chunk-%v.dat", cid)
	return path.Join(cs.dir, name)
}

func (cs *ChunkServer) GetChunk(cid int64, off int64, buf []byte) (int, error) {
	filename := cs.GetFileName(cid)

	fd, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()
	return fd.ReadAt(buf, off)
}

func (cs *ChunkServer) SetChunk(cid int64, off int64, buf []byte) (int, error) {
	if off+int64(len(buf)) > util.MAXCHUNKSIZE {
		log.Panic("chunk size cannot be larger than maxchunksize\n")
	}

	filename := cs.GetFileName(cid)

	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return 0, err
	}

	defer fd.Close()

	return fd.WriteAt(buf, off)
}

func (cs *ChunkServer) removeChunk(cid int64) error {
	filename := cs.GetFileName(cid)
	return os.Remove(filename)
}
