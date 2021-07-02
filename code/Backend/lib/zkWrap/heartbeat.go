package zkWrap

import (
	"github.com/deckarep/golang-set"
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type Heartbeat struct {
	conn			*zk.Conn
	timeout			time.Duration
	path			string
	closeChan		chan int
	mates			*[]string
	originMates		*[]string

	ServiceName		string
}

type HeartbeatEventCallback func(string, string)	// regData, nodeName

func RegisterHeartbeat(serviceName string, timeout time.Duration, regData string,
	onConnectCallback HeartbeatEventCallback, onDisConnectCallback HeartbeatEventCallback) (*Heartbeat, error) {

	if strings.ContainsRune(serviceName, '/') {
		return nil, errors.New("zkWrap: heartbeat service name cannot contain rune /")
	}

	conn, _, err := zk.Connect(hosts, timeout)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	path := pathWithChroot(heartbeatRoot + "/" + serviceName)

	if err := createContainerIfNotExist(conn, path); err != nil {
		return nil, errors.WithStack(err)
	}

	hbPath := path + "/" + regData
	for {
		if pExists, stat, err := conn.Exists(hbPath); err != nil {
			return nil, errors.WithStack(err)
		} else if pExists {
			if err := conn.Delete(hbPath, stat.Version); err != nil {
				if err == zk.ErrBadVersion {
					continue
				} else {
					return nil, errors.WithStack(err)
				}
			}
			break
		} else {
			break
		}
	}

	var mates, originMates []string
	var oldChildren, newChildren []string
	var evenChan <-chan zk.Event
	closeChan := make(chan int)
	if oldChildren, _, evenChan, err = conn.ChildrenW(path); err != nil {
		return nil, err
	} else {
		mateSet := mapset.NewSetFromSlice(stringSlice2InterfaceSlice(oldChildren))
		mateSet.Remove(regData)
		oldChildren = interfaceSlice2StringSlice(mateSet.ToSlice())
		originMates = make([]string, len(oldChildren)); copy(originMates, oldChildren)
		mates = oldChildren
		go func() {
			for {
				select {
				case event := <-evenChan:
					if event.Type == zk.EventNodeChildrenChanged {
						newChildren, _, evenChan, err = conn.ChildrenW(path)
						oldSet := mapset.NewSetFromSlice(stringSlice2InterfaceSlice(oldChildren))
						newSet := mapset.NewSetFromSlice(stringSlice2InterfaceSlice(newChildren))
						deleted := oldSet.Difference(newSet)
						added := newSet.Difference(oldSet)
						for del := range deleted.Iterator().C {
							onDisConnectCallback(regData, del.(string))
						}
						for add := range added.Iterator().C {
							onConnectCallback(regData, add.(string))
						}
						oldChildren = newChildren

						newSet.Remove(regData)
						mates = interfaceSlice2StringSlice(newSet.ToSlice())
					}
				case <-closeChan:
					return
				}
			}
		}()
	}

	newPath, err := conn.Create(hbPath, nil, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Heartbeat{
		conn: conn,
		timeout: timeout,
		path: newPath,
		closeChan: closeChan,
		mates: &mates,
		originMates: &originMates,

		ServiceName: serviceName,
	}, nil
}

func (hb *Heartbeat) Disconnect() {
	hb.closeChan <- 1
	hb.conn.Close()
}

func (hb *Heartbeat) GetMates() []string {
	return *hb.mates
}

func (hb *Heartbeat) GetOriginMates() []string {
	return *hb.originMates
}