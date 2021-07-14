package master

import (
	"DFS/util"
	"container/heap"
	"fmt"
	"math/rand"
	"sync"

	"github.com/sirupsen/logrus"
)

/* The state of all chunkServer, maintained by the master */

type ChunkServerStates struct {
	sync.RWMutex
	servers map[util.Address]*ChunkServerState
}

type ChunkServerState struct {
	sync.RWMutex
	// LastHeartbeat time.Time
	ChunkList []util.Handle
}
type ChunkServerHeap struct {
	Addr     util.Address
	ChunkNum int
}

// type serialChunkServerStates struct {
// 	Addr  util.Address
// 	State SerialChunkServerState
// }
// type SerialChunkServerState struct {
// 	LastHeartbeat time.Time
// 	ChunkList     []util.Handle
// }

type CssHeap []ChunkServerHeap

func (h CssHeap) Len() int           { return len(h) }
func (h CssHeap) Less(i, j int) bool { return h[i].ChunkNum < h[j].ChunkNum }
func (h CssHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *CssHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(ChunkServerHeap))
}

func (h *CssHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// func (s *ChunkServerStates) Serialize() []serialChunkServerStates {
// 	s.RLock()
// 	defer s.RUnlock()
// 	var scss = make([]serialChunkServerStates, 0)
// 	for key, state := range s.servers {
// 		state.RLock()
// 		scss = append(scss, serialChunkServerStates{Addr: key, State: SerialChunkServerState{
// 			ChunkList:     state.ChunkList,
// 			LastHeartbeat: state.LastHeartbeat,
// 		}})
// 		state.RUnlock()
// 	}
// 	return scss
// }
// func (s *ChunkServerStates) Deserialize(scss []serialChunkServerStates) error {
// 	s.Lock()
// 	defer s.Unlock()
// 	for _, _scss := range scss {
// 		s.servers[_scss.Addr] = &ChunkServerState{
// 			RWMutex:       sync.RWMutex{},
// 			LastHeartbeat: _scss.State.LastHeartbeat,
// 			ChunkList:     _scss.State.ChunkList,
// 		}
// 	}
// 	return nil
// }

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
	h := &CssHeap{}
	heap.Init(h)
	for addr, state := range s.servers {
		state.RLock()
		hea := ChunkServerHeap{ChunkNum: len(state.ChunkList), Addr: addr}
		heap.Push(h, hea)
		state.RUnlock()
	}
	for times > 0 {
		times--
		addrs = append(addrs, heap.Pop(h).(ChunkServerHeap).Addr)
	}
	logrus.Debugf("Balanced choosed ")
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
		// LastHeartbeat: time.Now(),
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
func (s *ChunkServerStates) addChunk(addr util.Address, handle util.Handle) error {
	s.servers[addr].Lock()
	defer s.servers[addr].Unlock()
	state := s.servers[addr]
	state.ChunkList = append(state.ChunkList, handle)
	return nil
}
func (s *ChunkServerStates) removeChunk(addr util.Address, handle util.Handle) error {
	s.servers[addr].Lock()
	defer s.servers[addr].Unlock()
	state := s.servers[addr]
	index := -1
	for _index, _handle := range state.ChunkList {
		if _handle == handle {
			index = _index
			break
		}
	}
	// unlikely
	if index == -1 {
		err := fmt.Errorf("removeChunk warning : %v is not existed in %v", handle, addr)
		return err
	}
	state.ChunkList = append(state.ChunkList[:index], state.ChunkList[index+1:]...)
	return nil
}

func (s *ChunkServerStates) GetServerHandleList(addr util.Address) []util.Handle {
	s.RLock()
	defer s.RUnlock()
	s.servers[addr].RLock()
	defer s.servers[addr].RUnlock()
	return s.servers[addr].ChunkList
}

func (s *ChunkServerStates) GetServerHandleOne(addr util.Address, index int) util.Handle {
	s.RLock()
	defer s.RUnlock()
	s.servers[addr].RLock()
	defer s.servers[addr].RUnlock()
	return s.servers[addr].ChunkList[index]
}

func newChunkServerState() *ChunkServerStates {
	ns := &ChunkServerStates{
		servers: make(map[util.Address]*ChunkServerState),
	}
	return ns
}
