package util

import "time"

// Master
type Handle int64
type DFSPath string
type LinuxPath string
type Address string
type Version int64

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
	INITIALVERSION    =  1
	HERETRYTIMES = 3
)

// ChunkServer

type ChunkState struct {
	Handle Handle
	VerNum Version
	//CheckSum int64
}

// Client
const (
	//MAXCHUNKSIZE = 64 << 20 // 64MB
	MAXCHUNKSIZE      = 1024 // 64B
	REPLICATIONTIMES  = 3
	MAXAPPENDSIZE     = MAXCHUNKSIZE / 2 // TODO: according to GFS docs, we should set it to MAXCHUNKSIZE / 4
	MAXFD             = 65535
	MINFD 			  = 1
	HEARTBEATDURATION = 2000 * time.Millisecond // 2s
	DELETEPREFIX      = "_delete_"
)

// Error Code
type ErrorCode int32

const (
	NOSPACE ErrorCode = -11
)
