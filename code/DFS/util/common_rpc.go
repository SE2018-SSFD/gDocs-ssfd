package util

type CacheID struct {
	Handle     Handle
	ClientAddr Address
}

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
type OpenArg struct {
	Path DFSPath `json:"path"`
}
type OpenRet struct {
	Fd int `json:"fd"`
}
type CloseArg struct {
	Fd int `json:"fd"`
}
type CloseRet struct {
}
type CreateArg struct {
	Path DFSPath `json:"path"`
}
type CreateRet struct {
}
type MkdirArg struct {
	Path DFSPath `json:"path"`
}
type MkdirRet struct {
}
type DeleteArg struct {
	Path DFSPath `json:"path"`
}
type DeleteRet struct {
}
type ListArg struct {
	Path DFSPath `json:"path"`
}
type ListRet struct {
	Files []string `json:"files"`
}
type WriteArg struct {
	Fd     int    `json:"fd"`
	Offset int    `json:"offset"`
	Data   []byte `json:"data"`
}
type WriteRet struct {
	BytesWritten int `json:"bytes_written"`
}
type GetReplicasArg struct {
	Path       DFSPath `json:"path"`
	ChunkIndex int     `json:"chunk_index"`
}
type GetReplicasRet struct {
	ChunkHandle      Handle    `json:"chunk_handle"`
	ChunkServerAddrs []Address `json:"chunk_server_addrs"`
}
type GetFileMetaArg struct {
	Path DFSPath `json:"path"`
}
type GetFileMetaRet struct {
	Exist bool `json:"exist"`
	IsDir bool `json:"is_dir"`
	Size  int  `json:"size"`
}
type SetFileMetaArg struct {
	Path DFSPath `json:"path"`
	Size int     `json:"size"`
}

type SetFileMetaRet struct {
}
