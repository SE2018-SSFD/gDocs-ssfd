package util

// Master
type Handle int64
type DFSPath string
type LinuxPath string
type Address string

// ChunkServer
type CacheID struct {
	Handle     Handle
	ClientAddr string
}

type ChunkState struct {
	Handle Handle
	VerNum int64
	//CheckSum int64
}

// Client
const (
	//MAXCHUNKSIZE = 64 << 20 // 64MB
	MAXCHUNKSIZE     = 64 // 64B
	REPLICATIONTIMES = 3
	MAXFD            = 128
)
