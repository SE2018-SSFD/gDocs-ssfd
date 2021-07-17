package master

import (
	"DFS/util"
	"github.com/sirupsen/logrus"
	"sort"
)

// Trigger by crontab or manual
func (m *Master) LoadBalanceCheck()(err error) {
	var retC util.CloneChunkReply
	var retS util.SetStaleReply
	logrus.Infof("LoadBalanceCheck finished")
	toSort:=make(CssHeap,0)

	avg := 0
	balanceCount := 0
	for addr,state := range m.css.servers{
		state.RLock()
		toSort = append(toSort,ChunkServerHeap{
			ChunkNum: len(state.ChunkList),
			Addr: addr,
		})
		avg += len(state.ChunkList)
		state.RUnlock()
	}
	avg = avg / len(m.css.servers)
	logrus.Debugln("avg :",avg, " toSort:",toSort)
	// load balance base limit
	if avg <= util.LBLIMIT{
		return
	}
	cssNum := len(toSort)
	if cssNum <= util.REPLICATIONTIMES{
		logrus.Fatal("NotEnoughServerError : not enough error in loadbalance\n")
		return
	}

	// Resort the array every time
	for{
		sort.Sort(toSort)
		provider := &toSort[cssNum-1]
		receiver := &toSort[0]
		//m.css.servers[provider.Addr].Lock()
		//m.css.servers[receiver.Addr].Lock()

		targetHandle := m.css.GetServerHandleOne(provider.Addr,0)
		if receiver.ChunkNum < avg*2/3{
			balanceCount += 1
			addrList := make([]util.Address,0)
			addrList = append(addrList,receiver.Addr)

			if provider.ChunkNum <= 0 {
				logrus.Warnf("LoadBalance failed bacause the provider has too few chunks")
				return
			}
			err = util.Call(string(provider.Addr), "ChunkServer.CloneChunkRPC", util.CloneChunkArgs{
				Addrs: addrList,
				Len:util.MAXCHUNKSIZE,
				Handle: targetHandle,
			}, &retC)
			if err!=nil{
				logrus.Warnf("LoadBalance failed of cloneChunk from %v to %v : %v",provider.Addr,receiver.Addr,err)
				return
			}
			staleHandleList := make([]util.Handle,0)
			staleHandleList = append(staleHandleList,targetHandle)
			err = util.Call(string(toSort[cssNum-1].Addr), "ChunkServer.SetStaleRPC", util.SetStaleArgs{Handles:staleHandleList },&retS)
			if err!=nil{
				logrus.Warnf("LoadBalance failed of SetStale from %v to %v : %v",provider.Addr,receiver.Addr,err)
				return
			}
			receiver.ChunkNum += 1
			provider.ChunkNum -= 1
			// Update master info
			err = m.cs.DeleteLocationOfChunk(provider.Addr,targetHandle)
			if err!=nil{
				logrus.Warnf("LoadBalance failed of delete location of addr %v in handle %v : %v",provider.Addr,targetHandle,err)
				return
			}
			err = m.cs.AddLocationOfChunk(receiver.Addr,targetHandle)
			if err!=nil{
				logrus.Warnf("LoadBalance failed of add location of addr %v in handle %v : %v",receiver.Addr,targetHandle,err)
				return
			}
			err = m.css.removeChunk(provider.Addr,targetHandle)
			if err!=nil{
				logrus.Warnf("LoadBalance failed of delete chunk of addr %v in handle %v : %v",provider.Addr,targetHandle,err)
				return
			}
			err = m.css.addChunk(receiver.Addr,targetHandle)
			if err!=nil{
				logrus.Warnf("LoadBalance failed of add chunk of addr %v in handle %v : %v",provider.Addr,targetHandle,err)
				return
			}
		}else{
			break
		}

	}
	logrus.Infof("LoadBalanceCheck finished, balance %v chunk",balanceCount)
	return
}

// For test only
func (m *Master) GetServersChunkNum()(result []ChunkServerHeap){
	m.css.RLock()
	defer m.css.RUnlock()
	logrus.Println("----1----")

	logrus.Println("remain servers:",m.css.servers)
	logrus.Println(m.cs.chunk)
	logrus.Println("----1----")

	for addr,state := range m.css.servers{
		state.RLock()
		result = append(result,ChunkServerHeap{
			ChunkNum: len(state.ChunkList),
			Addr: addr,
		})
		state.RUnlock()
	}
	return
}