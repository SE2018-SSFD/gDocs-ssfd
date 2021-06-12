package chunkserver

import "github.com/sirupsen/logrus"

func (cs *ChunkServer) Printf(format string, v ...interface{}) {
	var vv []interface{}
	vv = append(vv, cs.addr)
	vv = append(vv, v)
	logrus.Printf("chunkserver %v: "+format, vv...)
}

func (cs *ChunkServer) Crash() {
	err := cs.l.Close()
	close(cs.shutdown)
	if err != nil {
		return
	}
}
