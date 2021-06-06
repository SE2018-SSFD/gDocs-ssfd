package util
// Master
type Handle int64
type DFSPath string
type LinuxPath string
type Address string

// ChunkServer

// Client
const (
	MAXCHUNKSIZE = 64 << 20 // 64MB
	REPLICATIONTIMES = 3
)
