package master

import "time"

/* The state of all chunkServer, maintained by the master */


type ChunkServerStates struct {
	servers map[string]ChunkServerState
}

type ChunkServerState struct{
	lastHeartbeat time.Time
}