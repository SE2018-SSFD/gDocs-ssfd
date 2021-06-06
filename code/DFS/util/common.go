package util
// Master
type Handle string
type DFSPath string
type LinuxPath string
type Address string
// ChunkServer

// Client

// RPC structure
type CreateArg struct{
	Path DFSPath
}
type CreateRet struct {
	Result bool
}
type GetReplicasArg struct {
	chunkHandle Handle
}
type GetReplicasRet struct {
	ChunkServerAddrs []string
}