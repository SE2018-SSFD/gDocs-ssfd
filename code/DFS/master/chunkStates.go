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
	chunk map[util.Handle]*chunkState
	handle *handleState
}
type fileState struct{
	sync.RWMutex
	chunks []*chunkState
}
type chunkState struct {
	sync.RWMutex
	Handle util.Handle
	Locations []util.Address // set of replica locations
	Version util.Version
}
type handleState struct {
	sync.RWMutex
	curHandle util.Handle
}
type SerialfileState struct{
	Chunks []SerialChunkState
}
type SerialChunkStates struct{
	CurHandle util.Handle
	File  map[util.DFSPath]SerialfileState
}
type SerialChunkState struct{
	Handle util.Handle
	Version util.Version
}

// Serialize a chunkstates
func (s* ChunkStates) getChunkNum(path util.DFSPath) int{
	s.RLock()
	s.file[path].RLock()
	chunkNum := len(s.file[path].chunks)
	s.file[path].RUnlock()
	s.RUnlock()
	return chunkNum
}
func (s* ChunkStates) Serialize() SerialChunkStates {
	s.RLock()
	defer s.RUnlock()
	scss := SerialChunkStates{
		CurHandle: -1,
		File: make(map[util.DFSPath]SerialfileState),
	}
	for path,state := range s.file{
		s.file[path].RLock()
		chunks := make([]SerialChunkState,0)
		for index,chunk := range state.chunks{
			state.chunks[index].RLock()
			chunks = append(chunks,SerialChunkState{
				Handle: chunk.Handle,
				Version : chunk.Version,
			} )
			state.chunks[index].RUnlock()
		}
		scss.File[path]=SerialfileState{
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

	// recover original states
	for path,state := range scss.File{
		err := s.NewFile(path)
		if err!=nil{
			return err
		}
		for _,chunk := range state.Chunks{
			s.file[path].chunks = append(s.file[path].chunks,&chunkState{
				Handle: chunk.Handle,
			} )
		}
	}
	s.handle.curHandle = scss.CurHandle
	return nil
}
func (s* ChunkStates) AddLocationOfChunk(addr util.Address, handle util.Handle) error  {
	s.RLock()
	state := s.chunk[handle]
	state.Lock()
	defer state.Unlock()
	s.RUnlock()
	s.chunk[handle].Locations = append(s.chunk[handle].Locations,addr)
	return nil
}

func (s* ChunkStates) DeleteLocationOfChunk(addr util.Address,handle util.Handle) error {
	s.RLock()
	state := s.chunk[handle]
	state.Lock()
	defer state.Unlock()
	s.RUnlock()
	index := -1
	for _index,_addr := range state.Locations{
		if _addr == addr{
			index = _index
			break
		}
	}
	// unlikely
	if index == -1{
		err := fmt.Errorf("deleteLocationOfchunk error : %v is not existed in %v",handle,addr)
		return err
	}
	state.Locations = append(state.Locations[:index], state.Locations[index+1:]...)
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
	s.handle.Unlock()

	// add chunk to file
	newChunk = &chunkState{
		Locations: make([]util.Address,0),
		Handle: newHandle,
	}
	s.chunk[newHandle] = newChunk
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
		chunk: make(map[util.Handle]*chunkState,0),
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

