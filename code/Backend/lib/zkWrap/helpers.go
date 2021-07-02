package zkWrap

import (
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"sync"
)

var (
	BackSlashErr = errors.New("zkWrap: service name cannot contain rune /")
)

func pathWithChroot(path string) string {
	return root + path
}

func createContainerIfNotExist(conn *zk.Conn, path string) (err error) {
	for {
		if pExists, _, err := conn.Exists(path); err != nil {
			continue
		} else if pExists {
			return nil
		} else {
			for {
				_, err = conn.CreateContainer(path, nil, zk.FlagTTL, zk.WorldACL(zk.PermAll))
				if err != nil && err != zk.ErrNodeExists {
					return err
				} else {
					return nil
				}
			}
		}
	}
}

func deleteNodeOne(conn *zk.Conn, path string, version int32) (err error) {
	for {
		if err := conn.Delete(path, version); err != nil && err == zk.ErrNoNode {
			return errors.WithStack(err)
		} else if err != nil {
			continue
		} else {
			println(path)
			return nil
		}
	}
}

// deleteNodeAll is a recursive function which deletes all nodes in the path
func deleteNodeAll(conn *zk.Conn, path string, delSelf bool) (err error) {
	wg := sync.WaitGroup{}
	for {
		if children, stat, err := conn.Children(path); err != nil && err == zk.ErrNoNode {
			return errors.WithStack(err)
		} else if err != nil {
			continue
		} else {
			wg.Add(len(children))
			for _, child := range children {
				childPath := path + "/" + child
				go func() {
					deleteNodeAll(conn, childPath, true)
					wg.Done()
				}()
			}

			wg.Wait()
			if delSelf && stat.EphemeralOwner == 0 {
				if err := deleteNodeOne(conn, path, stat.Version); err != nil {
					return errors.WithStack(err)
				}
			}
			break
		}
	}

	return nil
}

func stringSlice2InterfaceSlice(before []string) (after []interface{}) {
	for _, v := range before {
		after = append(after, v)
	}
	return
}

func interfaceSlice2StringSlice(before []interface{}) (after []string) {
	for _, v := range before {
		after = append(after, v.(string))
	}
	return
}
