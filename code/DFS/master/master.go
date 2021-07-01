package master

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Master struct {
	sync.RWMutex
	logLock  sync.Mutex
	addr     util.Address
	metaPath util.LinuxPath
	L        net.Listener
	rpcs     *rpc.Server
	// manage the state of chunkserver node
	css *ChunkServerStates
	// manage the state of chunk
	cs               *ChunkStates
	ns               *NamespaceState
	shutdown         chan interface{}
	clusterHeartbeat *zkWrap.Heartbeat //this heartbeat contains one master and some chunkservers, when master become leader, join this heartbeat
	leaderHeartbeat  *zkWrap.Heartbeat //this heartbeat contains one master and some clients, when master become leader, join this heartbeat
	el               *zkWrap.Elector   // use for leader election
}

type OperationType int32

func InitMultiMaster(addr util.Address, metaPath util.LinuxPath) (*Master, error) {
	// Init RPC server
	m := &Master{
		addr:     addr,
		metaPath: metaPath,
		rpcs:     rpc.NewServer(),
		shutdown: make(chan interface{}),
	}
	err := m.rpcs.Register(m)
	if err != nil {
		logrus.Fatal("Register error:", err)
		os.Exit(1)
	}
	l, err := net.Listen("tcp", string(m.addr))
	if err != nil {
		logrus.Fatal("listen error:", err)
	}

	// Zookeeper connection
	var wg sync.WaitGroup // To sync all the goroutines
	wg.Add(util.MASTERCOUNT)

	onMasterConn := func(me string, who string) {
		logrus.Infoln(me + " receive :" + who)
		wg.Done()
	}
	onMasterDisConn := func(me string, who string) {
		logrus.Infoln(me + " receive disconn :" + who)
		wg.Add(1)
	}

	// Listening on other master
	hb, err := zkWrap.RegisterHeartbeat("master", util.MAXWAITINGTIME*time.Second, string(addr), onMasterConn, onMasterDisConn)
	if err != nil {
		return m, err
	}
	mate := len(hb.GetOriginMates())
	logrus.Debugln(string(addr) + ": has " + strconv.Itoa(mate) + "mate")
	for i := 0; i < mate; i++ {
		wg.Done()
	}
	m.L = l
	// Init zookeeper
	//c, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second) //*10)

	// Init metadata manager
	m.ns = newNamespaceState()
	m.cs = newChunkStates()
	m.css = newChunkServerState()

	// Create log file if not exist
	_, err = os.Stat(path.Join(string(m.metaPath), "log.dat"))
	if os.IsNotExist(err) {
		_, err = os.Create(path.Join(string(m.metaPath), "log.dat"))
		if err != nil {
			return m, err
		}
	}

	// Wait until other masters are ready
	// err = implicitWait(util.MAXWAITINGTIME*time.Second, &wg)

	//zookeeper election
	m.RegisterElectionNodes()

	// listening on chunkservers
	//cb := func(el *zkWrap.Elector) {
	//	onClientConn := func (me string, who string) {
	//		var argG util.GetChunkStatesArgs
	//		var retG util.GetChunkStatesReply
	//		err = util.Call(who, "Master.GetFileMetaRPC", argG, &retG)
	//		//TODO : check chunk states
	//		err = m.RegisterServer(util.Address(who))
	//		if err!=nil{
	//			logrus.Fatal("Master addServer error : ", err)
	//			return
	//		}
	//	}
	//	onClientDisConn := func (me string, who string) {
	//
	//	}
	//	hb,err = zkWrap.RegisterHeartbeat("addServers",util.MAXWAITINGTIME * time.Second,string(addr),onClientConn,onClientDisConn)
	//	if err !=nil {
	//		logrus.Fatal("Master addServer error : ", err)
	//		return
	//	}
	//}
	//_, err = zkWrap.NewElector("test", string(addr), cb)
	//if err !=nil {
	//	logrus.Fatal("Election error : ", err)
	//	return nil,err
	//}

	if err == nil {
		logrus.Infoln("master " + addr + ": init success")
	}
	return m, err
}
func InitMaster(addr util.Address, metaPath util.LinuxPath) (*Master, error) {
	// Init RPC server
	m := &Master{
		addr:     addr,
		metaPath: metaPath,
		rpcs:     rpc.NewServer(),
		shutdown: make(chan interface{}),
	}
	err := m.rpcs.Register(m)
	if err != nil {
		logrus.Fatal("Register error:", err)
		return m, err
	}
	l, err := net.Listen("tcp", string(m.addr))
	if err != nil {
		logrus.Fatal("listen error:", err)
		return m, err
	}
	logrus.Infoln("master " + addr + ": init success")
	m.L = l
	// Init zookeeper
	//c, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second) //*10)

	// Init metadata manager
	m.ns = newNamespaceState()
	m.cs = newChunkStates()
	m.css = newChunkServerState()
	err = m.TryRecover()
	if err != nil {
		logrus.Fatal("recover error:", err)
	}
	return m, err
}
func implicitWait(t time.Duration, wg *sync.WaitGroup) error {
	c := make(chan int)
	go func() {
		wg.Wait()
		c <- 0
	}()
	select {
	case <-c:
	case <-time.After(t):
		return fmt.Errorf("Parallel test failed\n")
	}
	return nil
}

func (m *Master) Serve() {
	// listening thread
	go func() {
		for {
			select {
			case <-m.shutdown:
				logrus.Debugln("Master shutdown!")
				return
			default:
			}
			conn, err := m.L.Accept()
			if err == nil {
				go func() {
					m.rpcs.ServeConn(conn)
					conn.Close()
				}()
			} else {

			}
		}
	}()
	//ch := make(chan bool)
	//<-ch
	//os.Exit(1)
}

// Direct Exit without storing the metadata
func (m *Master) Exit() {
	logrus.Debugf("Master Exit")
	err := m.L.Close()
	close(m.shutdown)
	// stop election
	m.el.StopElection()
	if err != nil {
		return
	}
}

func (m *Master) RegisterServer(addr util.Address) error {
	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.ADDSERVEROPS, Addr: addr})
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}
	err = m.css.registerServer(addr)
	return err
}

func (m *Master) UnregisterServer(addr util.Address) error {
	err := m.css.unRegisterServer(addr)
	return err
}

func (m *Master) GetStatusString() string {
	return "Master address :" + string(m.addr) + ",metaPath :" + string(m.metaPath)
}

// CreateRPC is called by client to create a new file
func (m *Master) CreateRPC(args util.CreateArg, reply *util.CreateRet) error {
	logrus.Debugf("RPC create, File Path : %s", args.Path)

	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.CREATEOPS, Path: args.Path})
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}

	// Modified metadata
	err = m.ns.Mknod(args.Path, false)
	if err != nil {
		logrus.Warnf("RPC create failed : %s\n", err)
		return err
	}
	err = m.cs.NewFile(args.Path)
	if err != nil {
		logrus.Warnf("RPC create failed : %s", err)
		return err
	}
	return nil
}

// MkdirRPC is called by client to create a new dir
func (m *Master) MkdirRPC(args util.MkdirArg, reply *util.MkdirRet) error {
	logrus.Debugf("RPC mkdir, Dir Path : %s", args.Path)

	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.MKDIROPS, Path: args.Path})
	if err != nil {
		logrus.Warnf("RPC mkdir failed : %s", err)
		return err
	}

	// Modified metadata
	err = m.ns.Mknod(args.Path, true)
	if err != nil {
		logrus.Warnf("RPC mkdir failed : %s", err)
		return err
	}
	return nil
}

// DeleteRPC is called by client to lazily delete a dir or file
func (m *Master) DeleteRPC(args util.DeleteArg, reply *util.DeleteRet) error {
	logrus.Debugf("RPC delete, Dir Path : %s", args.Path)

	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.DELETEOPS, Path: args.Path})
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}

	// Modified metadata
	err = m.cs.Delete(args.Path)
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}
	err = m.ns.Delete(args.Path)
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}
	return nil
}

// ListRPC is called by client to list content of a dir
func (m *Master) ListRPC(args util.ListArg, reply *util.ListRet) (err error) {
	logrus.Debugf("RPC list, Dir Path : %s", args.Path)
	reply.Files, err = m.ns.List(args.Path)
	if err != nil {
		logrus.Warnf("RPC list failed : %s", err)
	}
	return err
}

// ScanRPC is called by client to scan all file info of a dir
func (m *Master) ScanRPC(args util.ScanArg, reply *util.ScanRet) (err error) {
	logrus.Debugf("RPC list, Dir Path : %s", args.Path)
	files, err := m.ns.List(args.Path)
	if err != nil {
		logrus.Warnf("RPC list failed : %s", err)
	}
	reply.FileInfos = make([]util.GetFileMetaRet, 0)
	for _, file := range files {
		var ret util.GetFileMetaRet
		fullPath := path.Join(string(args.Path), file)
		err := m.GetFileMetaRPC(util.GetFileMetaArg{Path: util.DFSPath(fullPath)}, &ret)
		if err != nil {
			return err
		}
		reply.FileInfos = append(reply.FileInfos, ret)
	}
	return err
}

// GetFileMetaRPC retrieve the file metadata by path
func (m *Master) GetFileMetaRPC(args util.GetFileMetaArg, reply *util.GetFileMetaRet) error {
	logrus.Debugf("RPC getFileMeta, File Path : %s", args.Path)
	node, err := m.ns.GetNode(args.Path)
	if err != nil {
		logrus.Warnf("RPC getFileMeta failed : %s", err)
		*reply = util.GetFileMetaRet{
			Exist: false,
			IsDir: false,
			// Size: -1,
		}
		return err
	}
	reply.Exist = true
	reply.IsDir = node.isDir
	// if node.isDir{
	// 	reply.Size = -1
	// }else{
	// 	reply.Size = m.cs.file[args.Path].size
	// }
	return nil
}

// SetFileMetaRPC set the file metadata by path
func (m *Master) SetFileMetaRPC(args util.SetFileMetaArg, reply *util.SetFileMetaRet) error {
	logrus.Debugf("RPC setFileMeta, File Path : %s", args.Path)

	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.SETFILEMETAOPS, Path: args.Path, Size: args.Size})
	if err != nil {
		logrus.Warnf("RPC SetFileMeta failed : %s", err)
		return err
	}

	// Modified metadata
	m.cs.file[args.Path].Lock()
	defer m.cs.file[args.Path].Unlock()
	m.cs.file[args.Path].size = args.Size
	return nil
}

// GetReplicasRPC get a chunk handle by file path and offset
// as well as the addresses of servers which store the chunk (and its replicas)
// if offset == -1,return the last one
// TODO : add lease
func (m *Master) GetReplicasRPC(args util.GetReplicasArg, reply *util.GetReplicasRet) (err error) {
	// Check if file exist
	logrus.Debugf("RPC getReplica, file path : %s, chunk index : %d", args.Path, args.ChunkIndex)
	m.cs.RLock()
	fs, exist := m.cs.file[args.Path]
	if !exist {
		m.cs.RUnlock()
		err = fmt.Errorf("FileNotExistsError : the requested DFS path %s is non-existing", string(args.Path))
		return err
	}
	fs.Lock()
	m.cs.RUnlock()

	//if offset == -1 , return the last one
	if args.ChunkIndex == -1 {
		args.ChunkIndex = len(fs.chunks) - 1
		reply.ChunkIndex = len(fs.chunks) - 1
	}

	// Find the target chunk, if not exists, create one
	// Note that ChunkIndex <= len(fs.chunks) should be checked by client
	var targetChunk *chunkState
	if args.ChunkIndex == len(fs.chunks) {
		// randomly choose servers to store chunk replica
		var addrs []util.Address
		addrs, err = m.css.randomServers(util.REPLICATIONTIMES)
		if err != nil {
			return err
		}

		// Write ahead log
		err := m.AppendLog(MasterLog{OpType: util.GETREPLICASOPS, Path: args.Path, Addrs: addrs})
		if err != nil {
			logrus.Warnf("RPC SetFileMeta failed : %s\n", err)
			return err
		}

		// enter the function with write lock of fs
		targetChunk, err = m.cs.CreateChunkAndReplica(fs, addrs)
		if err != nil {
			return err
		}
		//TODO : Update ChunkServerState
		//m.css.xxx
	} else {
		fs.Unlock()
		targetChunk = fs.chunks[args.ChunkIndex]
	}
	logrus.Debugln("targetchunk : ", targetChunk)
	// Get target servers which store the replicate
	reply.ChunkServerAddrs = make([]util.Address, 0)
	for _, addr := range targetChunk.Locations {
		reply.ChunkServerAddrs = append(reply.ChunkServerAddrs, addr)
	}
	reply.ChunkHandle = targetChunk.Handle
	return nil
}
