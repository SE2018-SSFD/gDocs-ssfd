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
