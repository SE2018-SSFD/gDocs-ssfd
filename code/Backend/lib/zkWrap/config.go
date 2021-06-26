package zkWrap

import (
	"github.com/go-zookeeper/zk"
	"time"
)

var hosts = []string{
	"127.0.0.1:30086",
	"127.0.0.1:30087",
	"127.0.0.1:30088",
}

var root = "/"

const (
	sessionTimeout	= time.Second * 15
	heartbeatRoot	= "/heartbeat"
	lockRoot		= "/lock"
	electionRoot	= "/election"
)

func Chroot(path string) error {
	if path[len(path) - 1:] == "/" {
		path = path[0:len(path)-1]
	}

	conn, _, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return err
	}

	if rootExists, _, err := conn.Exists(path); err != nil {
		return err
	} else if !rootExists {
		if _, err := conn.CreateContainer(path, nil, zk.FlagTTL, zk.WorldACL(zk.PermAll)); err != nil {
			return err
		}
	}

	if lockRootExists, _, err := conn.Exists(path + lockRoot); err != nil {
		return err
	} else if !lockRootExists {
		if _, err := conn.CreateContainer(path + lockRoot, nil, zk.FlagTTL, zk.WorldACL(zk.PermAll)); err != nil {
			return err
		}
	}

	if heartbeatRootExists, _, err := conn.Exists(path + heartbeatRoot); err != nil {
		return err
	} else if !heartbeatRootExists {
		if _, err := conn.CreateContainer(path + heartbeatRoot, nil, zk.FlagTTL, zk.WorldACL(zk.PermAll)); err != nil {
			return err
		}
	}

	if electionRootExists, _, err := conn.Exists(path + electionRoot); err != nil {
		return err
	} else if !electionRootExists {
		if _, err := conn.CreateContainer(path + electionRoot, nil, zk.FlagTTL, zk.WorldACL(zk.PermAll)); err != nil {
			return err
		}
	}


	root = path

	return nil
}

func CurRoot() string {
	return root
}
