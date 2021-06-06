package master

import (
	"DFS/util"
	"time"
)
/* The state of all chunks, maintained by the master */

type ChunkStates struct {
	chunk map[util.Handle]*chunkState
	file  map[string]*fileState
}

type chunkState struct {
	location []util.Address // set of replica locations
	primary  util.Address   // primary chunkServer address
	expire   time.Time           // lease expire time
}
type fileState struct{

}