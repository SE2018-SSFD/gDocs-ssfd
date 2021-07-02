package chunkserver

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"time"
)

func onHeartbeatConn(_ string, who string) {

}

func onHeartbeatDisConn(_ string, who string) {

}

func (cs *ChunkServer) RegisterNodes() {
	count := util.HERETRYTIMES
	for count>0{
		hb, err := zkWrap.RegisterHeartbeat(
			"heartbeat",
			util.HEARTBEATDURATION,
			"chunkserver"+cs.addr,
			onHeartbeatConn,
			onHeartbeatDisConn,
		)
		if err != nil {
			time.Sleep(1*time.Second)
			count--
		}else{
			cs.clusterHeartbeat = hb
			for _, mate := range hb.GetOriginMates() {
				onHeartbeatConn("", mate)
			}
			break
		}
	}
}
