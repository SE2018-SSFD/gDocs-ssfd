package master

import (
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)
/* The state of all chunks, maintained by the master */

type ChunkStates struct {
	chunk map[util.Handle]*chunkState
	file  map[util.DFSPath]*fileState
	curHandle util.Handle
}

type chunkState struct {
	locations []util.Address // set of replica locations
	expire   time.Time           // lease expire time
}
type fileState struct{
	chunks []util.Handle
	size int
}


// CreateChunkAndReplica create metadata of a chunk and its replicas
// then it ask chunkservers to create chunks in Linux File System
func (s* ChunkStates) CreateChunkAndReplica(path util.DFSPath,addrs []util.Address) (newHandle util.Handle,err error) {
	newHandle = s.curHandle+1
	logrus.Infof(" CreateChunkAndReplica : new Handle %d\n",newHandle)
	s.curHandle+=1
	_,exist := s.file[path]
	if !exist{
		err = fmt.Errorf("UnexpectedError : file meta not exists in chunk states")
		return
	}
	s.file[path].chunks = append(s.file[path].chunks,newHandle)
	s.chunk[newHandle] = &chunkState{
		locations: make([]util.Address,0),
	}
	for _ , addr := range addrs{
		//err := util.Call(string(addr), "ChunkServer.RPCCreateChunk", gfs.CreateChunkArg{handle}, &r)
		//
		s.chunk[newHandle].locations = append(s.chunk[newHandle].locations,addr)
	}
	return
}

func newChunkStates()*ChunkStates{
	cs := &ChunkStates{
		chunk: make(map[util.Handle]*chunkState,0),
		file : make(map[util.DFSPath]*fileState,0),
	}
	return cs
}

// NewFile init the file metadata
func (s* ChunkStates) NewFile(path util.DFSPath) error {
	_,exist := s.file[path]
	if exist{
		return fmt.Errorf("UnexpectedError : file meta exists in chunk states\n")
	}
	s.file[path] = &fileState{
		chunks : make([]util.Handle,0),
		size:0,
	}
	return nil
}

// Delete a file and its chunks
func (s* ChunkStates) Delete(path util.DFSPath) error {
	fs,exist := s.file[path]
	if !exist{
		return fmt.Errorf("DeleteError : path %s is not existed",path)
	}
	for _,chunk := range fs.chunks{
		delete(s.chunk,chunk)
	}
	delete(s.file,path)
	return nil
}
