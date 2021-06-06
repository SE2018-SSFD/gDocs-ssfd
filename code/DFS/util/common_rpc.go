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
