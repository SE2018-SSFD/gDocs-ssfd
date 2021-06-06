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
	Err error
}
type MkdirArg struct{
	Path DFSPath
}
type MkdirRet struct {
	Err error
}
type GetReplicasArg struct {
	chunkHandle Handle
}
type GetReplicasRet struct {
	ChunkServerAddrs []string
	Err error
}
