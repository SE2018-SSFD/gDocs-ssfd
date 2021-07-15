package chunkserver

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func (cs *ChunkServer) StartRPCServer() error {
	rpcs := rpc.NewServer()
	rpcs.Register(cs)
	listener, err := net.Listen("tcp", string(cs.addr))
	if err != nil {
		log.Fatalf("Chunkserver listen %v error\n", string(cs.addr))
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
			}
		}
		log.Print("ChunkServer: done\n")
	}()

	zkWrap.Chroot("/DFS")
	time.Sleep(500 * time.Millisecond)
	cs.RegisterNodes()
	// cs.Printf("Register zookeeper node\n")
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

func (cs *ChunkServer) SetStaleRPC(args util.SetStaleArgs, reply *util.SetStaleReply) error {
	cs.Lock()
	for _, h := range args.Handles {
		cs.AppendLog(ChunkInfoLog{Handle: h, VerNum: 0, Operation: Operation_Delete})
		cs.RemoveChunk(h)
		cs.chunks[h].isStale = true
	}
	cs.Unlock()
	return nil
}

func (cs *ChunkServer) LoadDataRPC(args util.LoadDataArgs, reply *util.LoadDataReply) error {
	//log.Printf("ChunkServer %v: load data \n", cs.addr)
	cs.cache.Set(args.CID, args.Data)
	if len(args.Addrs) > 0 {
		newArgs := util.LoadDataArgs{
			Data:  args.Data,
			CID:   args.CID,
			Addrs: args.Addrs[1:],
		}
		err := util.Call(string(args.Addrs[0]), "ChunkServer.LoadDataRPC", newArgs, nil)
		if err != nil {
			//log.Panicf("ChunkServer %v: "+err.Error(), cs.addr)
			logrus.Debug(err)
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
	if err != nil && !strings.Contains(err.Error(),"EOF") {
		log.Fatalf("get chunk error : %v", err)
	}

	reply.Buf = buf[:len]
	reply.Len = len
	reply.VerNum = ck.verNum
	if args.Len != len {
		logrus.Printf("ChunkServer %v: read chunk len %v,but actual len %v\n", cs.addr, args.Len, len)
		// return fmt.Errorf("ChunkServer %v: read chunk len %v,but actual len %v", cs.addr, args.Len, len)
	}
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
	cs.AppendLog(ChunkInfoLog{Handle: args.Handle, VerNum: util.INITIALVERSION, Operation: Operation_Update})
	cs.chunks[args.Handle] = &ChunkInfo{verNum: util.INITIALVERSION}
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
	var pad bool = false
	if args.IsAppend {
		off, err := cs.AppendChunk(args.CID.Handle, data)
		logrus.Print("Append handle ", args.CID.Handle, " len ", len(data), " off ", off)
		if err != nil {
			logrus.Print(err)
		}
		if off == util.MAXCHUNKSIZE {
			reply.ErrorCode = util.NOSPACE
			pad = true // pad other chunkServer
		} else {
			reply.ErrorCode = 0
		}
		args.Off = off
		reply.Off = off
	} else {
		reply.ErrorCode = 0
		len, err := cs.SetChunk(args.CID.Handle, args.Off, data)
		logrus.Print("Handle ", args.CID.Handle, " len ", len, " off ", args.Off)
		if err != nil {
			logrus.Panic(err)
		}
	}

	ch := make(chan error)
	for _, secondaryAddr := range args.Addrs {
		go func(addr util.Address) {
			ch <- util.Call(string(addr), "ChunkServer.StoreDataRPC",
				util.StoreDataArgs{
					CID: args.CID,
					Off: args.Off,
					Pad: pad,
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
	var len int
	if args.Pad {
		// data = []byte{0}
		// len, err = cs.SetChunk(args.CID.Handle, util.MAXCHUNKSIZE-1, data)
		err = cs.PadChunk(args.CID.Handle)
	} else {
		len, err = cs.SetChunk(args.CID.Handle, args.Off, data)
	}
	log.Printf("ChunkServer %v: store handle %v, len %v\n", cs.addr, args.CID.Handle, len)
	return err
}

func (cs *ChunkServer) UpdateVersionRPC(args util.UpdateVersionArg, reply * util.UpdateVersionRet) error {
	cs.RLock()
	ck, exist := cs.chunks[args.Handle]
	if !exist {
		cs.RUnlock()
		return fmt.Errorf("ChunkServer %v: chunk %v not exist", cs.addr, args.Handle)
	}
	ck.RLock()
	cs.RUnlock()
	defer ck.RUnlock()
	if ck.verNum == args.Version {
		logrus.Print("Update chunk ",args.Handle," to version ",args.Version)
		cs.AppendLog(ChunkInfoLog{Handle: args.Handle, VerNum: ck.verNum + 1, Operation: Operation_Update})
		ck.verNum++
	}else {
		logrus.Print("Update chunk failed, chunk version is ",ck.verNum," ,master version is ",args.Version)
	}
	return nil
}

func (cs *ChunkServer) CloneChunkRPC(args util.CloneChunkArgs, reply *util.CloneChunkReply) error {
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

	len, err := cs.GetChunk(args.Handle, 0, buf)
	if err != nil && err != io.EOF{
		log.Fatalf("get chunk error\n")
		return err
	}

	//if args.Len != len {
	//	return fmt.Errorf("ChunkServer %v: clone chunk len %v,but actual len %v", cs.addr, args.Len, len)
	//}
	buf = buf[:len]

	err = util.CallAll(args.Addrs, "ChunkServer.SyncChunkRPC", util.SyncChunkArgs{Handle: args.Handle, VerNum: cs.chunks[args.Handle].verNum, Data: buf})
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (cs *ChunkServer) SyncChunkRPC(args util.SyncChunkArgs, reply *util.SyncChunkReply) error {
	cs.RLock()
	_, exist := cs.chunks[args.Handle]
	if !exist {
		cs.chunks[args.Handle] = &ChunkInfo{verNum: args.VerNum, isStale: false}
	}
	cs.chunks[args.Handle].Lock()
	cs.RUnlock()

	cs.chunks[args.Handle].isStale = false
	cs.chunks[args.Handle].verNum = args.VerNum

	defer cs.chunks[args.Handle].Unlock()
	_, err := cs.CreateAndSetChunk(args.Handle, 0, args.Data)
	if err != nil {
		return err
	}
	return nil
	// if l != len(args.Data) {
	// 	return fmt.Errorf("ChunkServer %v: clone chunk len %v,but actual len %v", cs.addr, args.Len, len)
	// }
}
