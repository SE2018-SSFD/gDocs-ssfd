package util

type CacheID struct {
	Handle     Handle
	ClientAddr string
}

type LoadDataArgs struct {
	Data  []byte
	CID   CacheID
	Addrs []Address
}

type LoadDataReply struct {
}

type SyncArgs struct {
	Handle Handle
	Addrs  []Address
}

type SyncReply struct {
}

type CreateChunkArgs struct {
	Handle Handle
}

type CreateChunkReply struct {
}

type ReadChunkArgs struct {
	Handle Handle
	Off    int
	Len    int
}
type ReadChunkReply struct {
	Len int
	Buf []byte
}

// Master RPC structure
type OpenArg struct {
	Path DFSPath
}
type OpenRet struct {
	Fd int
}
type CloseArg struct {
	Fd int
}
type CloseRet struct {
}
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
type DeleteArg struct {
	Path DFSPath
}
type DeleteRet struct {
}
type ListArg struct {
	Path DFSPath
}
type ListRet struct {
	Files []string
}
type GetReplicasArg struct {
	Path       DFSPath
	ChunkIndex int64
}
type GetReplicasRet struct {
	ChunkHandle      Handle
	ChunkServerAddrs []Address
}
type GetFileMetaArg struct {
	Path       DFSPath
}
type GetFileMetaRet struct{
	Exist bool
	IsDir bool
	Size int32
}

