## ZooKeeper

### 实现功能

#### 心跳 & 服务发现

利用ZooKeeper中容器节点断开连接后自动删除的特性。如果有服务器崩溃，它创建的容器节点就会自动删除；新加入的节点会创建新的容器节点。这样可以通过节点的存在性来实现服务发现和心跳。

利用ZooKeeper中的*ChildrenWatch*获得当前目录下子节点名称，并注册回调事件，当子节点发生改变时触发。值得注意的是，这个注册函数是**一次性**的，也就是说每次回调函数触发时，原先注册的监听事件会被取消，于是需要重新注册。这样会带来新的问题：在回调事件**触发**和**重新注册**之间的事件会被遗漏（据我所知，在GitHub上许多Go语言的ZooKeeper库都有这个问题，它们只是简单地使用for & select在新事件来的时候重新注册）。为了解决这个问题，我们利用了*ChildrenWatch*在注册和获得子节点列表上的原子性。当回调事件触发时，我们不仅仅关注触发事件的节点，而是将重新调用*ChildrenWatch*后获得的新的子节点集合和之前的集合一起进行考虑。这样旧集合和新集合差集是断开连接的服务器，新集合和就集合的差集则是新加入的服务器。

[Code Snippet 1](#code1)

#### 消息队列

消息队列是以一个房间的形式组织的：

* Log Room
  * Log Channel 1
  * Log Channel 2
  * ...
  
加入Log Room的服务器会收到所有Log Channel里的新消息的回调和新的Log Channel被创建的回调。

该消息队列的实现主要是为了满足DFS中Master日志同步的需求，例如这样的场景：

1. 主Master新建一个房间，新建一个消息队列
2. 主Master往队列中写日志，副Master收到回调在本地同步
3. 主Master决定进行checkpoint，为了避免阻塞，新开一个消息队列进行同步
4. 副Master收到新消息队列的回调，开始监听新的队列
5. 主Master做好checkpoint，删除旧队列中的所有日志

[Code Snippet 2](#code2)

#### 选举

所有参与选举的服务器在某个节点下创建顺序容器节点，序号最小的被选举为Leader。其它的节点需要监听在这个节点上，当Leader主动Step Down或者崩溃时，最小节点将会易位，新的leader被选举成功。

我们对现有的库进行了封装。

[Code Snippet 3](#code3)

#### 互斥锁

通过创建同一个节点，创建成功的服务器获得锁；放锁时删除该节点。等待锁的服务器会监听事件，当节点被删除时重新竞争创建。

我们对现有的库进行了封装。

[Code Snippet 4](#code4)

### 测试

#### 心跳 & 服务发现

1. 为四个服务器创建回调函数，使用四个Map记录收到的对应节点事件的数量
2. 四个服务器依次注册，每次注册时调用*GetOriginMates*得到的伙伴数量应该和之前注册的服务器数量相同
3. 检查每个服务器收到的*onConn*事件数量是否正确
4. 每个服务器主动断开连接，检查每个服务器收到的*onDisConn*事件数量是否正确

#### 消息队列

1. 创建10个Log Room
2. 为每个Room创建20/5个初始/追加的Channel，测试初始Channel和追加事件是否正确
3. 为每个channel创建5/20个初始/追加的Log，测试初始Log和追加事件是否正确
4. 使用WaitGroup等待预计的事件结束，并设置超时时间，超时则测试失败

#### 选举

1. 生成10个候选者
2. 每个候选者成为Leader后立刻退位
3. 测试是否同时只有一个Leader，且退位后状态是否改变
4. 使用WaitGroup等待预计的事件结束，并设置超时时间，超时则测试失败

#### 互斥锁

1. 生成10个锁的竞争者
2. 每个竞争者竞争到锁后睡眠一段时间再放锁
3. 测试释放锁的一定是得到锁的
4. 拿到锁时给计数加1，放锁时减1
5. 测试最后计数是否为0

#### 测试结果

通过率：100%
覆盖率：72.6%

未覆盖的大多数都是为了健壮性添加的错误处理，只是简单地返回错误，不需要覆盖。

### 附录

#### 代码节选

##### Code Snippet 1

<div id="code1"></div>

```go
for {
    select {
    case event := <-evenChan:
        if event.Type == zk.EventNodeChildrenChanged {
            newChildren, _, evenChan, err = conn.ChildrenW(path)
            oldSet := mapset.NewSetFromSlice(stringSlice2InterfaceSlice(oldChildren))
            newSet := mapset.NewSetFromSlice(stringSlice2InterfaceSlice(newChildren))
            deleted := oldSet.Difference(newSet)    // old - new = deleted
            added := newSet.Difference(oldSet)      // new - old = added

            for del := range deleted.Iterator().C {
                go onDisConnectCallback(regData, del.(string))
            }
            for add := range added.Iterator().C {
                go onConnectCallback(regData, add.(string))
            }

            oldChildren = newChildren

            newSet.Remove(regData)
            mates = interfaceSlice2StringSlice(newSet.ToSlice())
        }
    case <-closeChan:
        return
    }
}
```

##### Code Snippet 2

<div id="code2"></div>

```go
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
```

#### Code Snippet 3

<div id="code3"></div>

```go
for {
    select {
        case status, ok := <-election.Status():
            if ok {
                if status.Err != nil {
                    elector.IsLeader = false
                    elector.IsRunning = false
                    return
                } else if status.Role == leaderelection.Leader {
                    elector.IsLeader = true
                    onElectedCallback(&elector)
                }
            }
        }
    }
}
```


#### Code Snippet 4

<div id="code4"></div>

```go
type Mutex struct {
    lock    *zk.Lock

    Name    string
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
```