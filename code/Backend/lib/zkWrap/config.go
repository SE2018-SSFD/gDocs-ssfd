package zkWrap

import (
	"backend/utils/config"
	"github.com/go-zookeeper/zk"
	"strings"
	"time"
)

var hosts []string

var root = "/"

const (
	sessionTimeout	= time.Second * 15
	heartbeatRoot	= "/heartbeat"
	lockRoot		= "/lock"
	electionRoot	= "/election"
	logRoot			= "/log"
)

func Chroot(path string) error {
	hosts = strings.Split(config.Get().ZKAddr, ";")

	if path[len(path) - 1:] == "/" {
		path = path[0:len(path)-1]
	}

	conn, _, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return err
	}

	if err := createContainerIfNotExist(conn, path); err != nil {
		return err
	}

	if err := createContainerIfNotExist(conn, path + lockRoot); err != nil {
		return err
	}

	if err := createContainerIfNotExist(conn, path + heartbeatRoot); err != nil {
		return err
	}

	if err := createContainerIfNotExist(conn, path + electionRoot); err != nil {
		return err
	}

	if err := createContainerIfNotExist(conn, path + logRoot); err != nil {
		return err
	}

	root = path

	return nil
}

func CurRoot() string {
	return root
}
