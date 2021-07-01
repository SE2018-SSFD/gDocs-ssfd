package zkWrap

import (
	"github.com/Rossil2012/go-leaderelection"
	"github.com/go-zookeeper/zk"
)

type Elector struct {
	elect *leaderelection.Election
	conn  *zk.Conn

	IsLeader     bool
	IsRunning    bool
	ElectionName string
	Me           string
}

type ElectionCallback func(*Elector)

func NewElector(electionName string, me string, onElectedCallback ElectionCallback) (*Elector, error) {
	conn, _, err := zk.Connect(hosts, sessionTimeout/3)
	if err != nil {
		return nil, err
	}

	path := pathWithChroot(electionRoot + "/" + electionName)

	if pExists, _, err := conn.Exists(path); err != nil {
		return nil, err
	} else if !pExists {
		if _, err := conn.CreateContainer(path, nil, zk.FlagTTL, zk.WorldACL(zk.PermAll)); err != nil {
			return nil, err
		}
	}

	election, err := leaderelection.NewElection(conn, path, me)
	if err != nil {
		return nil, err
	}

	elector := Elector{
		elect: election,
		conn:  conn,

		IsLeader:     false,
		IsRunning:    true,
		ElectionName: electionName,
		Me:           me,
	}
	go election.ElectLeader()
	go func() {
		for {
			select {
			case status, ok := <-election.Status():
				if ok {
					if status.Err != nil {
						elector.IsLeader = false
						elector.IsRunning = false
					} else if status.Role == leaderelection.Leader {
						elector.IsLeader = true
						onElectedCallback(&elector)
					}
				} else {

				}
			}
		}
	}()

	return &elector, nil
}

func (el *Elector) Resign() {
	el.elect.Resign()
	el.IsLeader = false
}

func (el *Elector) StopElection() {
	// el.elect.EndElection()
	el.conn.Close()
	el.IsLeader = false
	el.IsRunning = false
}
