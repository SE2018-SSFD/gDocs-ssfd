package master

import (
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/rpc"
	"os"
)

type Master struct {
	addr    string
	 string
	metaPath string
	l          net.Listener
	rpcs       *rpc.Server
}

func InitMaster(addr string, metaPath string) *Master{
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
	return m
}

func (m *Master)Serve(){
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
	return "Master address :"+m.addr+ ",metaPath :"+m.metaPath
}