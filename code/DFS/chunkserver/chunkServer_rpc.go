package chunkserver

import (
	"DFS/util"
	"fmt"
	"log"
	"net"
	"net/rpc"
)

func (cs *ChunkServer) StartRPCServer() error {
	rpcs := rpc.NewServer()
	rpcs.Register(cs)
	listener, err := net.Listen("tcp", string(cs.addr))
	if err != nil {
		log.Fatalf("ChunKserver listen %v error\n", string(cs.addr))
	}

	cs.l = listener

	//listening
	go func() {
	loop:
		for {
			select {
			case <-cs.shutdown:
				break loop
			default:
			}
			conn, err := cs.l.Accept()
			if err == nil {
				go func() {
					rpcs.ServeConn(conn)
					conn.Close()
				}()
			} else {
				log.Fatal("Chunkserver: accept error\n")
				break
			}
		}
		log.Fatal("ChunkServer: done\n")
	}()

	//
	// go func() {
	// 	hbTicker := time.Tick(util.HEARTBEATDURATION)
	// loop:
	// 	for {
	// 		select {
	// 		case <- cs.shutdown:
	// 			break loop
	// 		case <- hbTicker:
	// 			err := cs.
	// 		}
	// 	}
	// }
	return err
}

func (cs *ChunkServer) GetChunkStatesRPC(args util.GetChunkStatesArgs, reply *util.GetChunkStatesReply) error {
	var chunkStates []util.ChunkState
	for handle, chunk := range cs.chunks {
		chunkStates = append(chunkStates, util.ChunkState{Handle: handle, VerNum: chunk.verNum})
	}
	reply.ChunkStates = chunkStates
	return nil
}

func (cs *ChunkServer) LoadDataRPC(args util.LoadDataArgs, reply *util.LoadDataReply) error {
	log.Printf("ChunkServer %v: load data \n", cs.addr)
	cs.cache.Set(args.CID, args.Data)
	if len(args.Addrs) > 0 {
		newArgs := util.LoadDataArgs{
			Data:  args.Data,
			CID:   args.CID,
			Addrs: args.Addrs[1:],
		}
		err := util.Call(string(args.Addrs[0]), "ChunkServer.LoadDataRPC", newArgs, nil)
		if err != nil {
			log.Panicf("ChunkServer %v: "+err.Error(), cs.addr)
		}
		return err
	}
	return nil
}

func (cs *ChunkServer) ReadChunkRPC(args util.ReadChunkArgs, reply *util.ReadChunkReply) error {
	buf := make([]byte, args.Len)

	cs.RLock()
	ck, exist := cs.chunks[args.Handle]
	if !exist {
		cs.RUnlock()
		return fmt.Errorf("ChunkServer %v: chunk %v not exist", cs.addr, args.Handle)
	}
	ck.RLock()
	cs.RUnlock()
	defer ck.RUnlock()

	len, err := cs.GetChunk(args.Handle, args.Off, buf)
	if err != nil {
		log.Fatalf("get chunk error\n")
	}

	reply.Buf = buf[:len]
	reply.Len = len
	reply.VerNum = ck.verNum
	if args.Len != len {
		return fmt.Errorf("ChunkServer %v: read chunk len %v,but actual len %v", cs.addr, args.Len, len)
	}
	return nil
}

func (cs *ChunkServer) HeartBeatRPC(args util.HeartBeatArgs, reply *util.HeartBeatReply) error {
	return nil
}

func (cs *ChunkServer) CreateChunkRPC(args util.CreateChunkArgs, reply *util.CreateChunkReply) error {
	log.Printf("ChunkServer %v: create chunk %v\n", cs.addr, args.Handle)
	cs.Lock()
	defer cs.Unlock()
	if _, ok := cs.chunks[args.Handle]; ok {
		log.Panicf("ChunkServer %v: chunk %v has been already created", cs.addr, args.Handle)
		return nil
	}
	cs.chunks[args.Handle] = &ChunkInfo{verNum: 0}
	return cs.CreateChunk(args.Handle)
}

//TODO : Append log to log file
//call by client
func (cs *ChunkServer) SyncRPC(args util.SyncArgs, reply *util.SyncReply) error {

	data, err := cs.cache.GetAndRemove(args.CID)
	if err != nil {
		return err
	}

	cs.RLock()
	ck, exist := cs.chunks[args.CID.Handle]
	if !exist {
		cs.RUnlock()
		return fmt.Errorf("ChunkServer %v: chunk %v not exist", cs.addr, args.CID.Handle)
	}
	ck.Lock()
	cs.RUnlock()
	defer ck.Unlock()

	cs.SetChunk(args.CID.Handle, args.Off, data)
	ch := make(chan error)
	for _, secondaryAddr := range args.Addrs {
		go func(addr util.Address) {
			ch <- util.Call(string(addr), "ChunkServer.StoreDataRPC",
				util.StoreDataArgs{
					CID: args.CID,
					Off: args.Off,
				}, nil)
		}(secondaryAddr)
	}

	for range args.Addrs {
		err := <-ch
		errs := ""
		if err != nil {
			log.Fatal(err)
			errs += err.Error() + "\n"
		}
		if errs != "" {
			return fmt.Errorf(errs)
		}
	}
	// TODO: error handler?

	return nil
}

func (cs *ChunkServer) StoreDataRPC(args util.StoreDataArgs, reply *util.StoreDataReply) error {
	data, err := cs.cache.GetAndRemove(args.CID)
	if err != nil {
		return err
	}

	cs.RLock()
	ck, exist := cs.chunks[args.CID.Handle]
	if !exist {
		cs.RUnlock()
		return fmt.Errorf("ChunkServer %v: chunk %v not exist", cs.addr, args.CID.Handle)
	}
	ck.Lock()
	cs.RUnlock()
	defer ck.Unlock()

	len, err := cs.SetChunk(args.CID.Handle, args.Off, data)
	log.Printf("ChunkServer %v: store handle %v, len %v\n", cs.addr, args.CID.Handle, len)
	return err
}
