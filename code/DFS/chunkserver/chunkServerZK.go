package chunkserver

import (
	"DFS/util"
	"DFS/util/zkWrap"
)

func onHeartbeatConn(_ string, who string) {

}

func onHeartbeatDisConn(_ string, who string) {

}

func (cs *ChunkServer) RegisterNodes() {
	hb, err := zkWrap.RegisterHeartbeat(
		"heartbeat",
		util.HEARTBEATDURATION,
		"chunkserver"+cs.addr,
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
