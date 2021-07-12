package master

import (
	"DFS/kafka"
	"DFS/util"
	"DFS/util/zkWrap"
	"github.com/sirupsen/logrus"
	"strings"
)

func (m *Master) onClusterHeartbeatConn(_ string, who string) {
	//listening on chunkservers
	if strings.Compare("chunkserver", who[:11]) == 0 {
		//TODO: check and update chunk states
		chunkServerAddr := util.Address(who[11:])
		logrus.Infof("chunkserver %v exists", chunkServerAddr)
		// Get chunk states in chunkServer, retry until RPCServer is available
		var argG util.GetChunkStatesArgs
		var retG util.GetChunkStatesReply
		count := util.HERETRYTIMES
		for count > 0 {
			err := util.Call(string(chunkServerAddr), "ChunkServer.GetChunkStatesRPC", argG, &retG)
			if err != nil {
				logrus.Warnf("Master addServer error : %v,retry", err)
				count--
			} else {
				break
			}
		}
		if count == 0 {
			logrus.Fatal("Master addServer error : cannot connect to chunkServer RPC  ")
		}

		m.RLock()
		defer m.RUnlock()
		staleHandles := make([]util.Handle, 0)
		for _, chunk := range retG.ChunkStates {
			ver, exist := m.cs.chunk[chunk.Handle]
			if !exist || chunk.VerNum != ver {
				if chunk.VerNum < ver {
					logrus.Infof("chunk %v with version %v is outdated", chunk.Handle, chunk.VerNum)
				} else {
					logrus.Fatalf("chunk %v with version %v is unexpected! Check the implementation", chunk.Handle, chunk.VerNum)
				}
				staleHandles = append(staleHandles, chunk.Handle)
			}
		}
		// SendBack stale chunk to chunkServer
		var argS util.SetStaleArgs
		var retS util.SetStaleReply
		argS.Handles = staleHandles
		err := util.Call(string(chunkServerAddr), "ChunkServer.SetStaleRPC", argS, &retS)
		if err != nil {
			logrus.Fatal("Master addServer error : ", err)
			return
		}
		err = m.RegisterServer(chunkServerAddr)
		if err != nil {
			logrus.Fatal("Master addServer error : ", err)
			return
		}
		logrus.Info("Master addServer success: ", who)

	}
}

func (m *Master) onClusterHeartbeatDisConn(_ string, who string) {
	if strings.Compare("chunkserver", who[:11]) == 0 {
		//TODO: remove chunkserver state
		chunkServerAddr := util.Address(who[11:])
		err := m.UnregisterServer(chunkServerAddr)
		if err != nil {
			logrus.Fatal("Master removeServer error : ", err)
			return
		}
		logrus.Infof("chunkserver %v leaves", chunkServerAddr)
	}
}

func (m *Master) onLeaderHeartbeatConn(_ string, who string) {
}

func (m *Master) onLeaderHeartbeatDisConn(_ string, who string) {
}

// RegisterClusterNodes is called by main master to deal with chunkServer
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
		//kafka producer
		m.ap, err = kafka.MakeProducer(string(m.addr))
		if err != nil {
			logrus.Fatal("kafka make producer error :", err)
		}
		m.RegisterClusterNodes()
		m.RegisterLeaderNodes()
	}
	m.el, err = zkWrap.NewElector("MasterLeaderElection", string(m.addr), cb)
	if err != nil {
		logrus.Fatal("new elector error : ", err)
		return
	}
}
