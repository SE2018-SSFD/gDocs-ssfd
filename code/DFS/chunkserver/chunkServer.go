package chunkserver

import "net"

type ChunkServer struct {
	addr    string
	masterAddr    string
	string
	dataPath string
	l          net.Listener
}

func InitChunkServer(chunkAddr string, dataPath string,masterAddr string) *ChunkServer{
	m := &ChunkServer{
		addr: chunkAddr,
		dataPath: dataPath,
		masterAddr : masterAddr,
	}
	return m
}
func (m *ChunkServer)GetStatusString()string{
	return "ChunkServer address :"+m.addr+ ",dataPath :"+m.dataPath
}