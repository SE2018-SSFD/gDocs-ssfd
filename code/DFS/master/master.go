package master

import (
	"DFS/util"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/rpc"
	"os"
)

type Master struct {
	addr    util.Address
	metaPath util.LinuxPath
	l          net.Listener
	rpcs       *rpc.Server
	// manage the state of chunkserver node
	css        *ChunkServerStates
	// manage the state of chunk
	cs		   *ChunkStates
	ns		   *NamespaceState
}

func InitMaster(addr util.Address, metaPath util.LinuxPath) *Master{
	// Init RPC server
	m := &Master{
		addr: addr,
		metaPath: metaPath,
		rpcs : rpc.NewServer(),
	}
	err := m.rpcs.Register(m)
	if err != nil {
		logrus.Fatal("Register error:",err)
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
	return m
}

func (m *Master)Serve(){
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

func (m *Master)GetStatusString()string{
	return "Master address :"+string(m.addr)+ ",metaPath :"+string(m.metaPath)
}

// CreateRPC is called by client to create a new file
func (m *Master) CreateRPC(args util.CreateArg, reply *util.CreateRet) error {
	logrus.Debugf("RPC create, File Path : %s\n",args.Path)
	err := m.ns.Mknod(args.Path,false)
	if err!=nil{
		logrus.Debugf("RPC create failed : %s\n",err)
	}else{
		logrus.Debugf("RPC create succeed\n")
	}
	return err
}

// MkdirRPC is called by client to create a new dir
func (m *Master) MkdirRPC(args util.MkdirArg, reply *util.MkdirRet) error {
	logrus.Debugf("RPC mkdir, Dir Path : %s\n",args.Path)
	err := m.ns.Mknod(args.Path,true)
	if err!=nil{
		logrus.Debugf("RPC mkdir failed : %s\n",err)
	}else{
		logrus.Debugf("RPC mkdir succeed\n")
	}
	return err
}

// ListRPC is called by client to list content of a dir
func (m *Master) ListRPC(args util.ListArg, reply *util.ListRet) (err error) {
	logrus.Debugf("RPC list, Dir Path : %s\n",args.Path)
	reply.Files,err = m.ns.List(args.Path)
	if err!=nil{
		logrus.Debugf("RPC list failed : %s\n",err)
	}else{
		logrus.Debugf("RPC list succeed\n")
	}
	return err
}
