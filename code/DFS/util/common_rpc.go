package util

type LoadDataArgs struct {
	Data   []byte
	Handle Handle
	Addrs  []Address
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
