package zkWrap

import (
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"strings"
)

type Mutex struct {
	lock	*zk.Lock

	Name	string
}

func NewMutex(lockName string) (*Mutex, error) {
	if strings.ContainsRune(lockName, '/') {
		return nil, errors.WithStack(BackSlashErr)
	}

	conn, _, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	println(pathWithChroot(lockRoot+"/"+lockName))
	return &Mutex{
		lock: zk.NewLock(conn, pathWithChroot(lockRoot+"/"+lockName), zk.WorldACL(zk.PermAll)),
		Name: lockName,
	}, nil
}

func (l *Mutex) Lock() error {
	return l.lock.Lock()
}

func (l *Mutex) Unlock() error {
	return l.lock.Unlock()
}
