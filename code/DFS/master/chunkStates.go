package master

import (
	"DFS/util"
	"time"
)
/* The state of all chunks, maintained by the master */

type ChunkStates struct {
	chunk map[util.Handle]*chunkState
	file  map[util.DFSPath]*fileState
}

type chunkState struct {
	locations []util.Address // set of replica locations
	expire   time.Time           // lease expire time
}
type fileState struct{
	chunks []util.Handle
}

func newChunkStates()*ChunkStates{
	cs := &ChunkStates{
		chunk: make(map[util.Handle]*chunkState,0),
		file : make(map[util.DFSPath]*fileState,0),
	}
	return cs
}