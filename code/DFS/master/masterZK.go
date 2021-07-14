package master

import (
	"DFS/kafka"
	"DFS/util"
	"DFS/util/zkWrap"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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
		// err := m.RegisterServer(chunkServerAddr)
		// if err != nil {
		// 	logrus.Fatal("Master addServer error : ", err)
		// 	return
		// }
		// check chunk version & add location to chunk
		m.css.servers[chunkServerAddr] = &ChunkServerState{} // clear all
		m.css.servers[chunkServerAddr].Lock()
		defer m.css.servers[chunkServerAddr].Unlock()
		staleHandles := make([]util.Handle, 0)
		m.cs.RLock()
		for _, chunk := range retG.ChunkStates {
			state, exist := m.cs.chunk[chunk.Handle]
			if !exist || chunk.VerNum != state.Version {
				if chunk.VerNum < state.Version {
					logrus.Infof("chunk %v with version %v is outdated", chunk.Handle, chunk.VerNum)
				} else {
					logrus.Fatalf("chunk %v with version %v is unexpected! Check the implementation", chunk.Handle, chunk.VerNum)
				}
				staleHandles = append(staleHandles, chunk.Handle)
			} else {
				m.cs.AddLocationOfChunk(chunkServerAddr, chunk.Handle)
				m.css.servers[chunkServerAddr].ChunkList = append(m.css.servers[chunkServerAddr].ChunkList, chunk.Handle)
			}
		}
		m.cs.RUnlock()

		// SendBack stale chunk to chunkServer
		var argS util.SetStaleArgs
		var retS util.SetStaleReply
		argS.Handles = staleHandles
		err := util.Call(string(chunkServerAddr), "ChunkServer.SetStaleRPC", argS, &retS)
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
		var err error
		for _, handle := range m.GetHandleList(chunkServerAddr) {
			err = m.DeleteLocationOfChunk(chunkServerAddr, handle)
			if err != nil {
				logrus.Fatal("Master removeLocation error : ", err)
				return
			}
		}
		// err = m.UnregisterServer(chunkServerAddr)
		m.css.Lock()
		defer m.css.Unlock()
		delete(m.css.servers, chunkServerAddr)
		// if err != nil {
		// 	logrus.Fatal("Master removeServer error : ", err)
		// 	return
		// }
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
		// sleep 1 second,wait all log write finish
		time.Sleep(1 * time.Second)
		err = m.TryRecover()
		if err != nil {
			logrus.Fatal("recover error:", err)
		}
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
