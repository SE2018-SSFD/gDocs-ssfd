package master

import (
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)
/* The state of all chunks, maintained by the master */

type ChunkStates struct {
	sync.RWMutex
	file  map[util.DFSPath]*fileState
	handle *handleState
}

type chunkState struct {
	sync.RWMutex
	handle util.Handle
	locations []util.Address // set of replica locations
	expire   time.Time           // lease expire time
}
type fileState struct{
	sync.RWMutex
	chunks []*chunkState
	size int
}
type handleState struct {
	sync.RWMutex
	curHandle util.Handle
}


// CreateChunkAndReplica create metadata of a chunk and its replicas
// then it ask chunkservers to create chunks in Linux File System
// Note : s.file[path] is locked now
func (s* ChunkStates) CreateChunkAndReplica(fs *fileState,addrs []util.Address) (newHandle util.Handle,err error) {
	var arg util.CreateChunkArgs
	var ret util.CreateChunkReply

	// increment handle
	s.handle.Lock()
	newHandle = s.handle.curHandle+1
	logrus.Infof(" CreateChunkAndReplica : new Handle %d\n",newHandle)
	s.handle.curHandle+=1
	s.handle.Unlock()

	// add chunk to file
	newChunk := &chunkState{
		locations: make([]util.Address,0),
		handle: newHandle,
		expire: time.Now(),
	}
	fs.chunks = append(fs.chunks,newChunk)
	newChunk.Lock()
	defer newChunk.Unlock()
	fs.Unlock()
	for _ , addr := range addrs{
		arg.Handle = newHandle
		err = util.Call(string(addr), "ChunkServer.CreateChunkRPC", arg, &ret)
		if err != nil {
			return 0, err
		}
		fs.chunks[newHandle].locations = append(fs.chunks[newHandle].locations,addr)
	}
	return
}

func newChunkStates()*ChunkStates{
	cs := &ChunkStates{
		file : make(map[util.DFSPath]*fileState,0),
		handle: &handleState{
			curHandle: 0,
		},
	}
	return cs
}

// NewFile init the file metadata
func (s* ChunkStates) NewFile(path util.DFSPath) error {
	s.Lock()
	defer s.Unlock()
	_,exist := s.file[path]
	if exist{
		return fmt.Errorf("UnexpectedError : file meta exists in chunk states\n")
	}
	s.file[path] = &fileState{
		chunks : make([]*chunkState,0),
		size:0,
	}
	return nil
}

// Delete a file and its chunks, used in garbage collection
//func (s* ChunkStates) Delete(path util.DFSPath) error {
//	s.Lock()
//	defer s.Unlock()
//	fs,exist := s.file[path]
//	if !exist{
//		return fmt.Errorf("DeleteError : path %s is not existed",path)
//	}
//	for _,chunk := range fs.chunks{
//		delete(s.chunk,chunk)
//	}
//	delete(s.file,path)
//	return nil
//}

func (s* ChunkStates) Delete(path util.DFSPath) error {
	s.Lock()
	defer s.Unlock()
	_, filename, err := util.ParsePath(path)
	if err != nil {
		return err
	}
	node := s.file[path]
	delete(s.file,path)
	s.file[util.DFSPath(util.DELETEPREFIX+string(filename))] = node
	return err
}