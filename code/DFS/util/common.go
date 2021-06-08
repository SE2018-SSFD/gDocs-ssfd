package util
// Master
type Handle int64
type DFSPath string
type LinuxPath string
type Address string

// ChunkServer

// Client
const (
	//MAXCHUNKSIZE = 64 << 20 // 64MB
	MAXCHUNKSIZE = 64 // 64B
	REPLICATIONTIMES = 3
	MAXFD = 128
)
