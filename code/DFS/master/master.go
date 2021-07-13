package master

import (
	"DFS/kafka"
	"DFS/util"
	"DFS/util/zkWrap"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
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
	clusterHeartbeat *zkWrap.Heartbeat     //this heartbeat contains one master and some chunkservers, when master become leader, join this heartbeat
	leaderHeartbeat  *zkWrap.Heartbeat     //this heartbeat contains one master and some clients, when master become leader, join this heartbeat
	el               *zkWrap.Elector       // use for leader election
	cg               *sarama.ConsumerGroup // kafka consumer group, use for logging
	ap               *sarama.AsyncProducer // kafka producer, use for logging
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
	// Init Kafka client
	//config := sarama.NewConfig()

	// Init metadata manager
	m.ns = newNamespaceState()
	m.cs = newChunkStates()
	m.css = newChunkServerState()

	_, err = os.Stat(string(m.metaPath))
	if err != nil {
		err := os.MkdirAll(string(m.metaPath), 0755)
		if err != nil {
			log.Fatalf("mkdir %v error\n", m.metaPath)
		}
	}
	// Create log file if not exist
	_, err = os.Stat(path.Join(string(m.metaPath), "log.dat"))
	if os.IsNotExist(err) {
		_, err = os.Create(path.Join(string(m.metaPath), "log.dat"))
		if err != nil {
			return m, err
		}
	}
	err = m.TryRecover()
	if err != nil {
		logrus.Fatal("recover error:", err)
	}

	// Wait until other masters are ready
	// err = implicitWait(util.MAXWAITINGTIME*time.Second, &wg)
	//kafka consumer
	m.cg, err = kafka.MakeConsumerGroup(string(m.addr))
	go kafka.Consume(m.cg, string(m.metaPath))

	//zookeeper election
	m.RegisterElectionNodes()
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
	// stop election (resign)
	//m.el.StopElection()
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

func (m *Master) GetHandleList(addr util.Address) []util.Handle {
	return m.css.GetServerHandleList(addr)
}

func (m *Master) DeleteLocationOfChunk(addr util.Address,handle util.Handle) error {
	return m.cs.DeleteLocationOfChunk(addr,handle)
}