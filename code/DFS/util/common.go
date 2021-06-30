package util

import "time"

// Master
type Handle int64
type DFSPath string
type LinuxPath string
type Address string

const (
	MASTERCOUNT    = 3
	MASTERZKPATH   = "/master"
	CREATEOPS      = 1000
	MKDIROPS       = 1001
	DELETEOPS      = 1002
	LISTOPS        = 1003
	GETFILEMETAOPS = 1004
	SETFILEMETAOPS = 1005
	GETREPLICASOPS = 1006
	ADDSERVEROPS   = 1007
	DELSERVEROPS   = 1008
)

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
	MAXAPPENDSIZE     = MAXCHUNKSIZE / 2
	MAXFD             = 128
	HEARTBEATDURATION = 200 * time.Millisecond
	DELETEPREFIX      = "_delete_"
)

// Error Code
type ErrorCode int32

const (
	NOSPACE ErrorCode = -11
)
