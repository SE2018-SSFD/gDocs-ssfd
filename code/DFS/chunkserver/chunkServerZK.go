package chunkserver

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"github.com/sirupsen/logrus"
	"strings"
)

func  (cs *ChunkServer)onHeartbeatConn(me string, who string) {
	if strings.Compare("master", who[:6]) == 0 {
		logrus.Infof("%v leader heartbeart conn : master leader %v join",me,who)
		cs.masterAddr = who[6:]
		//TODO: maybe we should clean fdTable
	}else{
		logrus.Infof("%v leader heartbeart conn : another chunkserver %v join",me,who)
	}
}

func onHeartbeatDisConn(_ string, who string) {

}

func (cs *ChunkServer) RegisterNodes() {
	hb, err := zkWrap.RegisterHeartbeat(
		"heartbeat",
		util.HEARTBEATDURATION,
		"chunkserver"+cs.addr,
		cs.onHeartbeatConn,
		onHeartbeatDisConn,
	)
	if err != nil {
		panic(err)
	}
	cs.clusterHeartbeat = hb
	for _, mate := range hb.GetOriginMates() {
		cs.onHeartbeatConn("", mate)
	}
}
