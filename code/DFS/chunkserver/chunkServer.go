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
	// size      int64
	sync.RWMutex
	verNum    int64 //version number
	mutations map[int64]*Mutation
	// checksum  int64
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

	log.Printf("chunkserver %v init success\n", chunkAddr)
	return cs
}

// func (cs *ChunkServer)

// func (cs *ChunkServer) HeartBeat() error {

// }

func (cs *ChunkServer) GetChunkStatesRPC(args util.GetChunkStatesArgs, reply *util.GetChunkStatesReply) error {
	var chunkStates []util.ChunkState
	for handle, chunk := range cs.chunks {
		chunkStates = append(chunkStates, util.ChunkState{Handle: handle, VerNum: chunk.verNum})
	}
	reply.ChunkStates = chunkStates
	return nil
}

func (cs *ChunkServer) GetStatusString() string {
	return "ChunkServer address :" + cs.addr + ",dir :" + cs.dir
}

func (cs *ChunkServer) GetFileName(handle util.Handle) string {
	name := fmt.Sprintf("chunk-%v.dat", handle)
	return path.Join(cs.dir, name)
}

func (cs *ChunkServer) GetChunk(handle util.Handle, off int, buf []byte) (int, error) {
	filename := cs.GetFileName(handle)

	fd, err := os.Open(filename)

	if err != nil {
		return 0, err
	}
	defer fd.Close()

	return fd.ReadAt(buf, int64(off))
}

// func (cs *ChunkServer) GetChunkWithOutLock(handle util.Handle, off int, buf []byte) (int, error) {
// 	filename := cs.GetFileName(handle)

// 	fd, err := os.Open(filename)

// 	if err != nil {
// 		return 0, err
// 	}
// 	defer fd.Close()

// 	return fd.ReadAt(buf, int64(off))
// }

func (cs *ChunkServer) SetChunk(handle util.Handle, off int, buf []byte) (int, error) {

	if off+len(buf) > util.MAXCHUNKSIZE {
		log.Panic("chunk size cannot be larger than maxchunksize\n")
	}

	filename := cs.GetFileName(handle)

	fd, err := os.OpenFile(filename, os.O_WRONLY, 0644)

	if err != nil {
		return 0, err
	}

	defer fd.Close()

	return fd.WriteAt(buf, int64(off))
}

func (cs *ChunkServer) RemoveChunk(handle util.Handle) error {
	filename := cs.GetFileName(handle)
	err := os.Remove(filename)
	if err != nil {
		log.Panic(err)
		return err
	}
	return nil
}

func (cs *ChunkServer) CreateChunk(handle util.Handle) error {
	filename := cs.GetFileName(handle)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
		return err
	}

	defer f.Close()
	err = f.Truncate(util.MAXCHUNKSIZE)
	if err != nil {
		log.Panic(err)
		return err
	}

	return nil
}
