package master

import (
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

/* The state of all chunkServer, maintained by the master */


type ChunkServerStates struct {
	servers map[util.Address]*ChunkServerState
}

// randomServers randomly choose times server from existing chunkservers
//goland:noinspection GoNilness
func (s *ChunkServerStates) randomServers(times int) (addrs []util.Address,err error) {
	// TODO:choose server in a load-balanced way
	if times > len(s.servers){
		err = fmt.Errorf("NotEnoughServerError : Not enough server to support %d times chunk replication\n",times)
		return
	}
	var all []util.Address
	for addr, _ := range s.servers {
		all = append(all, addr)
	}
	for _,serverIndex := range rand.Perm(len(s.servers))[:times]{
		addrs = append(addrs,all[serverIndex])
		logrus.Debugln(all[serverIndex]," ")
	}
	return
}

func (s *ChunkServerStates) RegisterServer(addr util.Address) error {
	_,exist := s.servers[addr]
	if exist{
		return fmt.Errorf("ServerReRegisterError : Server %s is registered\n",addr)
	}
	s.servers[addr] = &ChunkServerState{
		lastHeartbeat: time.Now(),
	}
	return nil
}

type ChunkServerState struct{
	lastHeartbeat time.Time
}

func newChunkServerState()*ChunkServerStates{
	ns := &ChunkServerStates{
		servers: make(map[util.Address]*ChunkServerState),
	}
	return ns
}
