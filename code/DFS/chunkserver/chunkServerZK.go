package chunkserver

import (
	"DFS/util/zkWrap"
	"time"
)

func onHeartbeatConn(_ string, who string) {

}

func onHeartbeatDisConn(_ string, who string) {

}

func (cs *ChunkServer) RegisterNodes() {
	hb, err := zkWrap.RegisterHeartbeat(
		"chunkserver",
		15*time.Second,
		cs.addr,
		onHeartbeatConn,
		onHeartbeatDisConn,
	)
	if err != nil {
		panic(err)
	}

	cs.clusterHeartbeat = hb

	for _, mate := range hb.GetOriginMates() {
		onHeartbeatConn("", mate)
	}
}
