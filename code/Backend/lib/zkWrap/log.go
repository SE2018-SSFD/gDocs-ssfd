package zkWrap

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
)

type LogOnNewChannelCallback func(cid int)

type LogRoom struct {
	conn		*zk.Conn
	name		string
	logs		map[int]*Log
	lock		sync.RWMutex
	cidSet		mapset.Set
	closeChan	chan int

	onAppendCallback		LogOnAppendCallback
	onNewChannelCallback	LogOnNewChannelCallback
}

func ClearLogRoom(serviceName string) (err error) {
	if strings.ContainsRune(serviceName, '/') {
		return errors.WithStack(BackSlashErr)
	}

	conn, _, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	path := pathWithChroot(logRoot + "/" + serviceName)

	if err = deleteNodeAll(conn, path, true); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func NewLogRoom(serviceName string, onNewChannelCallback LogOnNewChannelCallback, onAppendCallback LogOnAppendCallback,
	) (lr *LogRoom, originChanNum int, err error) {
	if strings.ContainsRune(serviceName, '/') {
		return nil, 0, errors.WithStack(BackSlashErr)
	}

	conn, _, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}

	path := pathWithChroot(logRoot + "/" + serviceName)

	if err := createContainerIfNotExist(conn, path); err != nil {
		return nil, 0, errors.WithStack(err)
	}

	var curChildren []string
	lr = &LogRoom{
		conn: conn,
		name: serviceName,
		logs: make(map[int]*Log),
		closeChan: make(chan int),
		onAppendCallback: onAppendCallback,
		onNewChannelCallback: onNewChannelCallback,
	}


	if originChildren, stat, evenChan, err := conn.ChildrenW(path); err == nil {
		originChanNum = int(stat.NumChildren)
		for i := 0; i < originChanNum; i += 1 {
			if err := lr.followLogChannel(i); err != nil {
				return nil, 0, err
			}
		}
		lr.cidSet = mapset.NewSetFromSlice(stringSlice2InterfaceSlice(originChildren))

		go func() {
			for {
				select {
				case event := <-evenChan:
					if event.Type == zk.EventNodeChildrenChanged {
						curChildren, stat, evenChan, err = conn.ChildrenW(path)
						if err == zk.ErrNoNode {
							return
						}
						curCidSet := mapset.NewSetFromSlice(stringSlice2InterfaceSlice(curChildren))
						diffCidSet := curCidSet.Difference(lr.cidSet)

						for chanName := range diffCidSet.Iterator().C {
							cid := name2id("logName", chanName.(string))
							if lr.GetLogChannel(cid) == nil {
								if err := lr.followLogChannel(cid); err != nil {
									return
								}
							}
							onNewChannelCallback(cid)
						}

						lr.cidSet = curCidSet
					}
				case <- lr.closeChan:
					return
				}
			}
		}()
	} else {
		return nil, 0, errors.WithStack(err)
	}

	return lr, originChanNum, nil
}

func (lr *LogRoom) NewLogChannel() (cid int, err error) {
	path := pathWithChroot(logRoot + "/" + lr.name + "/logRoom")
	path, err = lr.conn.Create(path, nil, zk.FlagSequence, zk.WorldACL(zk.PermAll))
	split := strings.Split(path, "/")
	cid = name2id("logRoom", split[len(split) - 1])
	chanName := lr.name + "/" + split[len(split) - 1]
	if err != nil {
		return 0, errors.WithStack(err)
	}

	lr.logs[cid] = &Log{}
	if l, err := NewLog(chanName, lr.onAppendCallback); err == nil {
		lr.lock.Lock()
		lr.logs[cid] = l
		lr.lock.Unlock()
	} else {
		return 0, errors.WithStack(err)
	}

	return cid, nil
}

func (lr *LogRoom) GetLogChannel(cid int) (l *Log) {
	lr.lock.RLock()
	l = lr.logs[cid]
	lr.lock.RUnlock()
	return l
}

func (lr *LogRoom) DisConnect() {
	lr.closeChan <- 1
	lr.conn.Close()
}


// followLogChannel joins a existed log channel
func (lr *LogRoom) followLogChannel(cid int) (err error) {
	chanName := lr.name + "/" + id2name("logRoom", cid)
	if l, err := NewLog(chanName, lr.onAppendCallback); err == nil {
		lr.lock.Lock()
		lr.logs[cid] = l
		lr.lock.Unlock()
		return nil
	} else {
		return errors.WithStack(err)
	}
}

type Log struct {
	conn			*zk.Conn
	path			string
	lidSet			mapset.Set
	originLogs		[]LogNode
	stopChan		chan int
}

type LogOnAppendCallback func(log LogNode)

type LogNode struct {
	Lid			int
	Content		string
}

// e.g. logName = "log0000000001"
func name2id(prefix string, name string) (id int) {
	if lid64, err := strconv.ParseInt(name[len(prefix):], 10, 64); err == nil {
		return int(lid64)
	} else {
		return 0
	}
}

func id2name(prefix string, id int) (name string) {
	const padding = 10
	idStr := strconv.Itoa(id)
	toPad := len(idStr)
	for i := 0; i < padding - toPad; i += 1 {
		idStr = "0" + idStr
	}
	return prefix + idStr
}

func NewLog(serviceName string, onAppendCallback LogOnAppendCallback) (l *Log, err error) {
	conn, _, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	path := pathWithChroot(logRoot + "/" + serviceName)

	if err := createContainerIfNotExist(conn, path); err != nil {
		return nil, errors.WithStack(err)
	}

	var curLogNames []string
	var originLogs []LogNode
	stopChan := make(chan int)

	retLog := Log{
		conn: conn,
		path: path,
		lidSet: mapset.NewSet(),
		stopChan: stopChan,
	}

	if originLogNames, _, evenChan, err := conn.ChildrenW(path); err == nil {
		for _, logName := range originLogNames {
			logPath := path + "/" + logName
			if logContent, _, err := conn.Get(logPath); err == nil {
				lid := name2id("log", logName)
				originLogs = append(originLogs, LogNode{
					Lid: lid,
					Content: string(logContent),
				})
				retLog.lidSet.Add(logName)
			} else {
				return nil, errors.WithStack(err)
			}
		}
		retLog.originLogs = originLogs
		go func() {
			for {
				select {
				case event := <-evenChan:
					if event.Type == zk.EventNodeChildrenChanged {
						curLogNames, _, evenChan, err = conn.ChildrenW(path)
						if err == zk.ErrNoNode {
							return
						}
						curLidSet := mapset.NewSetFromSlice(stringSlice2InterfaceSlice(curLogNames))
						newLogs := curLidSet.Difference(retLog.lidSet)
						for newLog := range newLogs.Iterator().C {
							logName := newLog.(string)
							logPath := path + "/" + logName
							retLog.lidSet.Add(logName)
							for {
								if content, _, err := conn.Get(logPath); err == nil {
									onAppendCallback(LogNode{
										Lid:     name2id("log", logName),
										Content: string(content),
									})
									break
								} else {
									continue
								}
							}
						}
					}
				case <-stopChan:
					return
				}
			}
		}()
	} else {
		return nil,  errors.WithStack(err)
	}

	return &retLog, nil
}

func (l *Log) Append(content string) (lid int, err error) {
	if path, err := l.conn.Create(l.path + "/log", []byte(content), zk.FlagSequence, zk.WorldACL(zk.PermAll)); err == nil {
		split := strings.Split(path, "/")
		return name2id("log", split[len(split) - 1]), nil
	} else {
		return 0, errors.WithStack(err)
	}
}

func (l *Log) DeleteAll() (err error) {
	if logNames, _, err := l.conn.Children(l.path); err == nil {
		for _, logName := range logNames {
			logPath := l.path + "/" + logName
			for {
				if exist, stat, err := l.conn.Exists(logPath); err == nil {
					if exist {
						if err := l.conn.Delete(logPath, stat.Version); err == zk.ErrNoNode {
							continue
						} else if err != nil {
							continue
						} else {
							break
						}
					} else {
						break
					}
				} else {
					continue
				}
			}
		}
		return nil
	} else {
		return errors.WithStack(err)
	}
}

func (l *Log) GetOriginLog() []LogNode {
	return l.originLogs
}

func (l *Log) DisConnect() {
	l.stopChan <- 1
	l.conn.Close()
}