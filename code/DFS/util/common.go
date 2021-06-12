package util

import "time"

// Master
type Handle int64
type DFSPath string
type LinuxPath string
type Address string

// ChunkServer

type ChunkState struct {
	Handle Handle
	VerNum int64
	//CheckSum int64
}

// Client
const (
	//MAXCHUNKSIZE = 64 << 20 // 64MB
	MAXCHUNKSIZE      = 64 // 64B
	REPLICATIONTIMES  = 3
	MAXFD             = 128
	HEARTBEATDURATION = 200 * time.Millisecond
	DELETEPREFIX = "_delete_"
)
