package master

import (
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)
/* The state of all chunks, maintained by the master */
/* Lock order : chunkstates > fileState > chunkState > handleState*/
type ChunkStates struct {
	sync.RWMutex
	file  map[util.DFSPath]*fileState
	chunk map[util.Handle]util.Version
	handle *handleState
}
type fileState struct{
	sync.RWMutex
	chunks []*chunkState
	size int
}
type chunkState struct {
	sync.RWMutex
	Handle util.Handle
	Locations []util.Address // set of replica locations
}
type handleState struct {
	sync.RWMutex
	curHandle util.Handle
}
type serialfileState struct{
	Size int
	Chunks []SerialChunkState
}
type SerialChunkStates struct{
	CurHandle util.Handle
	File  map[util.DFSPath]serialfileState
	Chunk map[util.Handle]util.Version
}
type SerialChunkState struct{
	Handle util.Handle
	Locations []util.Address // set of replica locations
}

// Serialize a chunkstates
func (s* ChunkStates) Serialize() SerialChunkStates {
	s.RLock()
	defer s.RUnlock()
	scss := SerialChunkStates{
		CurHandle: -1,
		File: make(map[util.DFSPath]serialfileState),
		Chunk: make(map[util.Handle]util.Version),
	}
	for handle,verNum := range s.chunk{
		scss.Chunk[handle]=verNum
	}
	for path,state := range s.file{
		s.file[path].RLock()
		chunks := make([]SerialChunkState,0)
		for index,chunk := range state.chunks{
			state.chunks[index].RLock()
			chunks = append(chunks,SerialChunkState{
				Locations: chunk.Locations,
				Handle: chunk.Handle,
			} )
			state.chunks[index].RUnlock()
		}
		scss.File[path]=serialfileState{
			Size : state.size,
			Chunks : chunks,
		}
		s.file[path].RUnlock()
	}
	s.handle.RLock()
	defer s.handle.RUnlock()
	scss.CurHandle = s.handle.curHandle
	return scss
}

// Deserialize into chunkstates
// Master need not take any lock because it is single-threaded
func (s* ChunkStates) Deserialize(scss SerialChunkStates) error {
	// clear remaining states
	s.file = make(map[util.DFSPath]*fileState)
	s.chunk = make(map[util.Handle]util.Version)

	// recover original states
	for handle,verNum := range scss.Chunk{
		s.chunk[handle]=verNum
	}
	for path,state := range scss.File{
		err := s.NewFile(path)
		if err!=nil{
			return err
		}
		s.file[path].size = state.Size
		for _,chunk := range state.Chunks{
			s.file[path].chunks = append(s.file[path].chunks,&chunkState{
				Locations: chunk.Locations,
				Handle: chunk.Handle,
			} )
		}
	}
	s.handle.curHandle = scss.CurHandle
	return nil
}


// CreateChunkAndReplica create metadata of a chunk and its replicas
// then it ask chunkservers to create chunks in Linux File System
// Note : s.file[path] is locked now
func (s* ChunkStates) CreateChunkAndReplica(fs *fileState,addrs []util.Address) (newChunk *chunkState,err error) {
	var arg util.CreateChunkArgs
	var ret util.CreateChunkReply

	// increment handle
	s.handle.Lock()
	newHandle := s.handle.curHandle+1
	logrus.Infof(" CreateChunkAndReplica : new Handle %d\n",newHandle)
	s.handle.curHandle+=1
	s.chunk[newHandle] = util.INITIALVERSION
	s.handle.Unlock()

	// add chunk to file
	newChunk = &chunkState{
		Locations: make([]util.Address,0),
		Handle: newHandle,
	}
	fs.chunks = append(fs.chunks,newChunk)

	newChunk.Lock()
	defer newChunk.Unlock()
	fs.Unlock()
	for _ , addr := range addrs{
		arg.Handle = newHandle
		err = util.Call(string(addr), "ChunkServer.CreateChunkRPC", arg, &ret)
		if err != nil {
			return nil, err
		}
		newChunk.Locations = append(newChunk.Locations,addr)
	}
	return
}

func newChunkStates()*ChunkStates{
	cs := &ChunkStates{
		file : make(map[util.DFSPath]*fileState,0),
		handle: &handleState{
			curHandle: 0,
		},
		chunk: make(map[util.Handle]util.Version,0),
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