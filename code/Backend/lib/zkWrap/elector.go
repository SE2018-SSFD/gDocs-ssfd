package zkWrap

import (
	"github.com/Rossil2012/go-leaderelection"
	"github.com/go-zookeeper/zk"
)

type Elector struct {
	elect			*leaderelection.Election

	IsLeader		bool
	IsRunning		bool
	ElectionName	string
	Me				string
}

type ElectionCallback func(*Elector)

func NewElector(electionName string, me string, onElectedCallback ElectionCallback) (*Elector, error) {
	conn, _, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return nil, err
	}

	path := pathWithChroot(electionRoot + "/" + electionName)

	if err := createContainerIfNotExist(conn, path); err != nil {
		return nil, err
	}

	election, err := leaderelection.NewElection(conn, path, me)
	if err != nil {
		return nil, err
	}

	elector := Elector{
		elect: election,

		IsLeader: false,
		IsRunning: true,
		ElectionName: electionName,
		Me: me,
	}
	go election.ElectLeader()
	go func() {
		for {
			select {
			case status, ok := <-election.Status():
				if ok {
					if status.Err != nil {
						election.EndElection()
						elector.IsLeader = false
						elector.IsRunning = false
					}
					if status.Role == leaderelection.Leader {
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
	el.elect.EndElection()
	el.IsLeader = false
	el.IsRunning = false
}