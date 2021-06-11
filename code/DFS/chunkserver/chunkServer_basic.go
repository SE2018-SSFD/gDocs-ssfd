package chunkserver

func (cs *ChunkServer) GetAddr() string {
	return string(cs.addr)
}
