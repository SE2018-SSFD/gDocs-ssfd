package master

import (
	"DFS/util"
	"container/heap"
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

type serialChunkServerStates struct {
	Addr  util.Address
	State ChunkServerState
}

type ChunkServerState struct {
	sync.RWMutex
	LastHeartbeat time.Time
	ChunkNum int
}
type ChunkServerHeap struct{
	Addr  util.Address
	ChunkNum int
}

type IntHeap []ChunkServerHeap

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i].ChunkNum < h[j].ChunkNum }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(ChunkServerHeap))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (s *ChunkServerStates) Serialize() []serialChunkServerStates {
	s.RLock()
	defer s.RUnlock()
	var scss = make([]serialChunkServerStates, 0)
	for key, value := range s.servers {
		scss = append(scss, serialChunkServerStates{Addr: key, State: *value})
	}
	return scss
}
func (s *ChunkServerStates) Deserialize(scss []serialChunkServerStates) error {
	s.Lock()
	defer s.Unlock()
	for _, _scss := range scss {
		s.servers[_scss.Addr] = &_scss.State
	}
	return nil
}

// randomServers randomly choose times server from existing chunkservers
//goland:noinspection GoNilness
func (s *ChunkServerStates) randomServers(times int) (addrs []util.Address, err error) {
	if times > len(s.servers) {
		err = fmt.Errorf("NotEnoughServerError : Not enough server to support %d times chunk replication\n", times)
		return
	}
	var all []util.Address
	for addr, _ := range s.servers {
		all = append(all, addr)
	}
	for _, serverIndex := range rand.Perm(len(s.servers))[:times] {
		addrs = append(addrs, all[serverIndex])
		//logrus.Debugln(all[serverIndex]," ")
	}
	return
}

// balanceServers choose server in a load-balanced way
//goland:noinspection GoNilness
func (s *ChunkServerStates) balanceServers(times int) (addrs []util.Address, err error) {
	if times > len(s.servers) {
		err = fmt.Errorf("NotEnoughServerError : Not enough server to support %d times chunk replication\n", times)
		return
	}
	h := &IntHeap{}
	heap.Init(h)
	for addr, state := range s.servers {
		hea := ChunkServerHeap{ChunkNum: state.ChunkNum,Addr: addr}
		heap.Push(h,&hea)
	}
	for times > 0{
		times --
		addrs = append(addrs,heap.Pop(h).(ChunkServerHeap).Addr)
	}

	return
}

func (s *ChunkServerStates) registerServer(addr util.Address) error {
	s.Lock()
	defer s.Unlock()
	_, exist := s.servers[addr]
	if exist {
		return fmt.Errorf("ServerRegisterError : Server %s is registered\n", addr)
	}
	s.servers[addr] = &ChunkServerState{
		LastHeartbeat: time.Now(),
	}
	return nil
}

func (s *ChunkServerStates) unRegisterServer(addr util.Address) error {
	s.Lock()
	defer s.Unlock()
	_, exist := s.servers[addr]
	if !exist {
		// return fmt.Errorf("ServerUnRegisterError : Server %s is not registered\n",addr)
		fmt.Printf("ServerUnRegisterError : Server %s is not registered\n", addr)
		return nil
	}
	delete(s.servers, addr)
	return nil
}

func newChunkServerState() *ChunkServerStates {
	ns := &ChunkServerStates{
		servers: make(map[util.Address]*ChunkServerState),
	}
	return ns
}
