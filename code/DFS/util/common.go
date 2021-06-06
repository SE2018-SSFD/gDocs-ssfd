package util

// Master
type Handle string
type DFSPath string
type LinuxPath string
type Address string

// ChunkServer

// Client

// RPC structure
type CreateArg struct {
	Path DFSPath
}
type CreateRet struct {
}
type MkdirArg struct {
	Path DFSPath
}
type MkdirRet struct {
}
type DeleteArg struct{
	Path DFSPath
}
type DeleteRet struct {
}
type ListArg struct{
	Path DFSPath
}
type ListRet struct {
	Files []string
}
type GetReplicasArg struct {
	chunkHandle Handle
}
type GetReplicasRet struct {
	ChunkServerAddrs []string
}

const (
	MAXCHUNKSIZE = 64 << 20 // 64MB
)
