package master

import (
	"DFS/util"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"github.com/sirupsen/logrus"
)

type Master struct {
	addr     util.Address
	metaPath util.LinuxPath
	l        net.Listener
	rpcs     *rpc.Server
	// manage the state of chunkserver node
	css *ChunkServerStates
	// manage the state of chunk
	cs *ChunkStates
	ns *NamespaceState
}

func InitMaster(addr util.Address, metaPath util.LinuxPath) *Master {
	// Init RPC server
	m := &Master{
		addr:     addr,
		metaPath: metaPath,
		rpcs:     rpc.NewServer(),
	}
	err := m.rpcs.Register(m)
	if err != nil {
		logrus.Fatal("Register error:", err)
		os.Exit(1)
	}
	l, err := net.Listen("tcp", string(m.addr))
	if err != nil {
		log.Fatal("listen error:", err)
	}
	m.l = l
	// Init zookeeper
	//c, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second) //*10)

	// Init metadata manager
	m.ns = newNamespaceState()
	m.cs = newChunkStates()
	m.css = newChunkServerState()
	return m
}

func (m *Master) Serve() {
	// listening thread
	go func() {
		for {
			conn, err := m.l.Accept()
			if err == nil {
				go func() {
					m.rpcs.ServeConn(conn)
					conn.Close()
				}()
			} else {

			}
		}
	}()
	// block on ch; make it a daemon
	ch := make(chan bool)
	<-ch
}

func (m*Master) RegisterServer(addr util.Address)error{
	err := m.css.RegisterServer(addr)
	return err
}

func (m *Master) GetStatusString() string {
	return "Master address :" + string(m.addr) + ",metaPath :" + string(m.metaPath)
}

// CreateRPC is called by client to create a new file
func (m *Master) CreateRPC(args util.CreateArg, reply *util.CreateRet) error {
	logrus.Debugf("RPC create, File Path : %s\n", args.Path)
	err := m.ns.Mknod(args.Path, false)
	if err != nil {
		logrus.Debugf("RPC create failed : %s\n", err)
		return err
	}
	err = m.cs.NewFile(args.Path)
	if err != nil {
		logrus.Debugf("RPC create failed : %s\n", err)
		return err
	}
	return nil
}

// MkdirRPC is called by client to create a new dir
func (m *Master) MkdirRPC(args util.MkdirArg, reply *util.MkdirRet) error {
	logrus.Debugf("RPC mkdir, Dir Path : %s\n", args.Path)
	err := m.ns.Mknod(args.Path, true)
	if err != nil {
		logrus.Debugf("RPC mkdir failed : %s\n", err)
		return err
	}
	return nil
}

// ListRPC is called by client to list content of a dir
func (m *Master) ListRPC(args util.ListArg, reply *util.ListRet) (err error) {
	logrus.Debugf("RPC list, Dir Path : %s\n", args.Path)
	reply.Files, err = m.ns.List(args.Path)
	if err != nil {
		logrus.Debugf("RPC list failed : %s\n", err)
	}
	return err
}

// GetFileMetaRPC retrieve the file metadata by path
func (m *Master) GetFileMetaRPC(args util.GetFileMetaArg, reply *util.GetFileMetaRet) error {
	logrus.Debugf("RPC getFileMeta, File Path : %s\n", args.Path)
	node,err := m.ns.GetNode(args.Path)
	if err != nil {
		logrus.Debugf("RPC getFileMeta failed : %s\n", err)
		reply = &util.GetFileMetaRet{
			Exist: false,
			IsDir: false,
			Size: -1,
		}
		return err
	}
	reply.Exist = true
	reply.IsDir = node.isDir
	if node.isDir{
		reply.Size = -1
	}else{
		reply.Size = m.cs.file[args.Path].size
	}
	return nil
}

// SetFileMetaRPC set the file metadata by path
func (m *Master) SetFileMetaRPC(args util.SetFileMetaArg, reply *util.SetFileMetaRet) error {
	logrus.Debugf("RPC getFileMeta, File Path : %s\n", args.Path)
	m.cs.file[args.Path].size = args.Size
	return nil
}

// GetReplicasRPC get a chunk handle by file path and offset
// as well as the addresses of servers which store the chunk (and its replicas)
func (m *Master) GetReplicasRPC(args util.GetReplicasArg, reply *util.GetReplicasRet) (err error) {
	// Check if file exist
	logrus.Debugf("RPC getReplica, file path : %s, chunk index : %d\n", args.Path, args.ChunkIndex)
	fs, exist := m.cs.file[args.Path]
	if !exist {
		err = fmt.Errorf("FileNotExistsError : the requested DFS path %s is non-existing!\n", string(args.Path))
		return err
	}

	// Find the target chunk, if not exists, create one
	// Note that ChunkIndex <= len(fs.chunks) should be checked by client
	var targetChunk util.Handle
	if int(args.ChunkIndex) == len(fs.chunks) {
		// randomly choose servers to store chunk replica
		var addrs []util.Address
		addrs,err = m.css.randomServers(util.REPLICATIONTIMES)
		if err!= nil{
			return err
		}
		targetChunk,err = m.cs.CreateChunkAndReplica(args.Path,addrs)
		//TODO : Update ChunkServerState
		//m.css.xxx
	} else {
		targetChunk = fs.chunks[args.ChunkIndex]
	}
	// Get target servers which store the replicate
	reply.ChunkServerAddrs = make([]util.Address, 0)
	for _, addr := range m.cs.chunk[targetChunk].locations {
		reply.ChunkServerAddrs = append(reply.ChunkServerAddrs, addr)
	}
	return nil
}
