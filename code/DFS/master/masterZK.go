package master

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"strings"

	"github.com/sirupsen/logrus"
)

func (m *Master) onClusterHeartbeatConn(_ string, who string) {
	//listening on chunkservers
	if strings.Compare("client", who[:6]) == 0 {
		//TODO: check and update chunk states
		clientAddr := util.Address(who[6:])
		logrus.Print("client %v exists", clientAddr)
	}
}

func (m *Master) onClusterHeartbeatDisConn(_ string, who string) {
	if strings.Compare("client", who[:6]) == 0 {
		//TODO: remove chunkserver state
		clientAddr := util.Address(who[6:])
		logrus.Print("client %v leaves", clientAddr)
	}
}

func (m *Master) onLeaderHeartbeatConn(_ string, who string) {
}

func (m *Master) onLeaderHeartbeatDisConn(_ string, who string) {
}

func (m *Master) RegisterClusterNodes() {
	hb, err := zkWrap.RegisterHeartbeat(
		"heartbeat",
		util.HEARTBEATDURATION,
		"master"+string(m.addr),
		m.onClusterHeartbeatConn,
		m.onClusterHeartbeatDisConn,
	)
	if err != nil {
		panic(err)
	}

	m.clusterHeartbeat = hb

	for _, mate := range hb.GetOriginMates() {
		m.onClusterHeartbeatConn("", mate)
	}
}

func (m *Master) RegisterLeaderNodes() {
	hb, err := zkWrap.RegisterHeartbeat(
		"masterLeader",
		util.HEARTBEATDURATION,
		"master"+string(m.addr),
		m.onLeaderHeartbeatConn,
		m.onLeaderHeartbeatDisConn,
	)
	if err != nil {
		panic(err)
	}

	m.clusterHeartbeat = hb

	for _, mate := range hb.GetOriginMates() {
		m.onClusterHeartbeatConn("", mate)
	}
}

func (m *Master) RegisterElectionNodes() {
	var err error
	cb := func(el *zkWrap.Elector) {
		//become leader, join heartbeat
		logrus.Print("master " + m.addr + " become leader!")
		m.RegisterClusterNodes()
		m.RegisterLeaderNodes()
	}
	m.el, err = zkWrap.NewElector("MasterLeaderElection", string(m.addr), cb)
	if err != nil {
		logrus.Fatal("new elector error")
		return
	}
}
