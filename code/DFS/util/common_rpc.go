package util

//ChunkServer RPC struct
type LoadDataArgs struct {
	Data  []byte
	CID   CacheID
	Addrs []Address
}

type LoadDataReply struct {
}

type StoreDataArgs struct {
	CID CacheID
	Off int
}

type StoreDataReply struct {
}

type SyncArgs struct {
	CID   CacheID
	Off   int
	Addrs []Address
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
	Len    int
	Buf    []byte
	VerNum int64
}

type GetChunkStatesArgs struct {
}

type GetChunkStatesReply struct {
	ChunkStates []ChunkState
}

type HeartBeatArgs struct {
}

type HeartBeatReply struct {
}

// Master RPC structure
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
	Path DFSPath
}
type GetFileMetaRet struct {
	Exist bool
	IsDir bool
	Size  int32
}
