package master

import (
	"DFS/util"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

/* The state of all chunkServer, maintained by the master */


type ChunkServerStates struct {
	sync.RWMutex
	servers map[util.Address]*ChunkServerState
}

type serialChunkServerStates struct{
	Addr util.Address
	State ChunkServerState
}

type ChunkServerState struct{
	LastHeartbeat time.Time
}
func (s *ChunkServerStates) Serialize() []serialChunkServerStates {
	s.RLock()
	defer s.RUnlock()
	var scss = make([]serialChunkServerStates,0)
	for key,value:= range s.servers{
		scss = append(scss,serialChunkServerStates{Addr: key,State: *value} )
	}
	return scss
}
func (s *ChunkServerStates) Deserialize(scss []serialChunkServerStates)error{
	s.Lock()
	defer s.Unlock()
	for _,_scss:= range scss{
		s.servers[_scss.Addr] = &_scss.State
	}
	return nil
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
		//logrus.Debugln(all[serverIndex]," ")
	}
	return
}

func (s *ChunkServerStates) RegisterServer(addr util.Address) error {
	s.Lock()
	defer s.Unlock()
	_,exist := s.servers[addr]
	if exist{
		return fmt.Errorf("ServerRegisterError : Server %s is registered\n",addr)
	}
	s.servers[addr] = &ChunkServerState{
		LastHeartbeat: time.Now(),
	}
	return nil
}

func (s *ChunkServerStates) UnRegisterServer(addr util.Address) error {
	s.Lock()
	defer s.Unlock()
	_,exist := s.servers[addr]
	if !exist{
		return fmt.Errorf("ServerUnRegisterError : Server %s is not registered\n",addr)
	}
	delete(s.servers,addr)
	return nil
}



func newChunkServerState()*ChunkServerStates{
	ns := &ChunkServerStates{
		servers: make(map[util.Address]*ChunkServerState),
	}
	return ns
}
