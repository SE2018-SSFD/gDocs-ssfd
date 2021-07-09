package service

import (
	"backend/dao"
	"backend/lib/cluster"
	"backend/lib/gdocFS"
	"backend/lib/wsWrap"
	"backend/lib/zkWrap"
	"backend/model"
	"backend/utils"
	"backend/utils/logger"
	"encoding/json"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	handleCellPoolSize	=		10000
	notifyChanSize		=		100
	stopChanSize		=		0		// to be reconsidered
)

var sheetWSMeta sync.Map	// fid -> sheetWSMetaEntry
var handleCellPool *ants.PoolWithFunc

func init() {
	var err error
	handleCellPool, err = ants.NewPoolWithFunc(handleCellPoolSize, handleCellPoolTask,
		ants.WithNonblocking(false))
	if err != nil {
		panic(err)
	}
}


type sheetWSMetaEntry struct {
	fid					uint
	cellMap				sync.Map				// cellKey{ row, col } -> *sheetWSCellMetaEntry
	userMap				sync.Map				// uid -> *sheetWSUserMetaEntry
	userN				int32					// number of connected users

	gLock				sync.RWMutex			// to lock events of specific fid
	cellWg				sync.WaitGroup			// waiting for handleCell to get all in-hand tasks done

	zkMutex				*zkWrap.Mutex			// distributed mutex to prevent service partition when adding new nodes
}

// wait - blocking until the state of specific fid becomes stable, all handleCell workers are blocking
// until leave is called
func (meta *sheetWSMetaEntry) wait() {
	meta.gLock.Lock()
	meta.cellWg.Wait()
}

// leave - allowing handleCell workers continue to process cell messages
func (meta *sheetWSMetaEntry) leave() {
	meta.gLock.Unlock()
}

type sheetWSUserMetaEntry struct {
	uid					uint
	username			string
	notifyChan			chan []byte		// notifyMessage -> notifyChan -> handleNotify
	stopChan			chan int
}

type sheetWSCellMetaEntry struct {
	cellKey				cellKey
	cellLock			cellLock
	workerSema			chan int	// semaphore(chan without buffer), guarantees only one worker is in critical section
	debugCnt			int32
}

type cellKey struct {
	Row		int
	Col		int
}

// client -> server

type sheetMessage struct {
	MsgType		string				`json:"msgType"`	// acquire, modify, release, onConn
	Body		json.RawMessage		`json:"body"`
}

type sheetAcquireMessage struct {
	Row			int 				`json:"row"`
	Col			int					`json:"col"`
}

type sheetModifyMessage struct {
	Row			int 				`json:"row"`
	Col			int					`json:"col"`
	Content		string				`json:"content"`
	Info		json.RawMessage		`json:"info"`
}

type sheetReleaseMessage struct {
	Row			int 				`json:"row"`
	Col			int					`json:"col"`
}

// server -> client

type sheetNotify struct {
	MsgType		string				`json:"msgType"`	// acquire, modify, release, onConn
	Body		json.RawMessage		`json:"body"`
}

type sheetPrepareNotify struct {
	Row			int 				`json:"row"`
	Col			int					`json:"col"`
	Username	string				`json:"username"`
}

type sheetModifyNotify struct {
	Row			int 				`json:"row"`
	Col			int					`json:"col"`
	Content		string				`json:"content"`
	Info		json.RawMessage		`json:"info"`
	Username	string				`json:"username"`
}

type sheetOnConnNotify struct {
	Name		string				`json:"name"`
	Columns		int					`json:"columns"`
	Content		[]string			`json:"content"`
	CellLocks	[]cellLockNotify	`json:"cellLocks"`
}

// controller.SheetBeforeUpgradeHandler

func SheetOnConnEstablished(token string, fid uint) (bool, int, *model.User, string) {
	if token == "" || fid == 0 {
		return false, utils.InvalidFormat, nil, ""
	}

	uid := CheckToken(token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, fid) {
			return false, utils.SheetNoPermission, nil, ""
		} else {
			sheet := dao.GetSheetByFid(fid)
			if sheet.IsDeleted {
				return false, utils.SheetIsInTrashBin, nil, ""
			}

			if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
				return false, utils.SheetWSRedirect, nil, addr
			}

			user := dao.GetUserByUid(uid)

			zkMutex, err := zkWrap.NewMutex("sheet" + strconv.Itoa(int(fid)))
			if err != nil {
				panic(err)
			}

			if actual, loaded := sheetWSMeta.LoadOrStore(fid, &sheetWSMetaEntry{
				fid: fid,
				zkMutex: zkMutex,
			}); loaded {
				meta := actual.(*sheetWSMetaEntry)
				if _, ok := meta.userMap.Load(uid); ok {
					return false, utils.SheetDupConnection, nil, ""
				}
			} else {	// get zkMutex for once
				logger.Infof("[fid(%d)\tuid(%d)] Trying to get ZK mutex!", fid, uid)
				// locking this file by fid for exclusive websocket service
				if err := zkMutex.Lock(); err != nil {
					panic(err)
				}
				logger.Infof("[fid(%d)\tuid(%d)] Get ZK mutex!", fid, uid)
			}

			return true, 0, &user, ""
		}
	} else {
		return false, utils.InvalidToken, nil, ""
	}
}


// WS Event Handlers: SheetOnConn, SheetOnDisConn, SheetOnMessage

func SheetOnConn(wss *wsWrap.WSServer, uid uint, username string, fid uint) {
	logger.Infof("[uid(%d)\tusername(%s)\tfid(%d)] Connected to server!", uid, username, fid)
	if v, ok := sheetWSMeta.Load(fid); !ok {
		logger.Errorf("![uid(%d)\tusername(%s)\tfid(%d)] No sheetWSMetaEntry on connection!", uid, username, fid)
	} else {
		userMeta := sheetWSUserMetaEntry{
			uid:      uid,
			username: username,
			notifyChan: make(chan []byte, notifyChanSize),
			stopChan: make(chan int, stopChanSize),
		}
		meta := v.(*sheetWSMetaEntry)

		meta.wait()
		defer meta.leave()		// TODO(flag): waiting

		logger.Infof("[uid(%d)\tusername(%s)\tfid(%d)] Get Lock onConn!", uid, username, fid)

		// send OnConn message
		sheet := dao.GetSheetByFid(fid)

		inCache := true
		memSheet := getSheetCache().Get(fid)
		// not in sheetCache, load from log and do eviction if needed
		if memSheet == nil {
			if memSheet, inCache = recoverSheetFromLog(sheet.Fid); memSheet == nil {
				panic("recoverSheetFromLog fails")
			}
		}

		memSheet.Lock()
		_, columns := memSheet.Shape()
		body := sheetOnConnNotify{
			Name: sheet.Name,
			Columns: columns,
			Content: memSheet.ToStringSlice(),
			CellLocks: dumpLocksOnCell(meta),
		}
		memSheet.Unlock()
		bodyRaw, _ := json.Marshal(body)
		sheetNtf := sheetNotify{
			MsgType: "onConn",
			Body: bodyRaw,
		}
		ntfRaw, _ := json.Marshal(sheetNtf)

		if inCache {		// TODO: if not in cache?
			keys, evicted := getSheetCache().Put(fid)
			commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
		}

		// leave to handleNotify
		userMeta.notifyChan <- ntfRaw

		// store in userMap
		if _, loaded := meta.userMap.LoadOrStore(uid, &userMeta); !loaded {
			logger.Infof("[userN(%d)\tfid(%d)] Current user on connection",
				atomic.AddInt32(&meta.userN, 1), fid)
			go handleNotify(wss, meta, &userMeta)
		} else {
			logger.Errorf("![uid(%d)\tusername(%s)\tfid(%d)] Duplicated connection!", uid, username, fid)
		}
	}
}

func SheetOnDisConn(wss *wsWrap.WSServer, uid uint, username string, fid uint) {
	logger.Infof("[uid(%d)\tusername(%s)\tfid(%d)] Disconnected from server!", uid, username, fid)

	if v, ok := sheetWSMeta.Load(fid); !ok {
		logger.Errorf("![uid(%d)\tusername(%s)\tfid(%d)] No sheetWSMetaEntry on disconnection!", uid, username, fid)
	} else {
		meta := v.(*sheetWSMetaEntry)

		meta.wait()
		defer meta.leave()

		if v, ok = meta.userMap.Load(uid); ok {
			userMeta := v.(*sheetWSUserMetaEntry)
			userMeta.stopChan <- 1
			meta.userMap.Delete(uid)
			meta.cellMap.Range(func(k interface{}, v interface{}) bool {
				cellMeta := v.(*sheetWSCellMetaEntry)
				ownerUid := uint(cellMeta.cellLock.owner)
				if ownerUid == userMeta.uid {
					cellMeta.cellLock.unLock(cellMeta.cellLock.owner)
				}
				return true
			})
		} else {
			logger.Errorf("![uid(%d)\tusername(%s)\tfid(%d)] No sheetWSUserMetaEntry on disconnection!",
				uid, username, fid)
		}

		if curUserN := atomic.AddInt32(&meta.userN, -1); curUserN == 0 {	// TODO(bug): may delete sheetWSMetaEntry after onConn
			logger.Infof("[uid(%d)\tusername(%s)\tfid(%d)\tuserN(%d)] Delete sheetWSMetaEntry!",
				uid, username, fid, curUserN)

			logger.Infof("[uid(%d)\tusername(%s)\tfid(%d)] Get Lock onDisConn!", uid, username, fid)

			// delete group entry
			sheetWSMeta.Delete(fid)

			// unlock zkMutex
			if err := meta.zkMutex.Unlock(); err != nil {
				logger.Errorf("[uid(%d)\tusername(%s)\tfid(%d)] Fail to unlock zkMutex!",uid, username, fid)
			}
			logger.Infof("[uid(%d)\tusername(%s)\tfid(%d)] Unlock zkMutex!", uid, username, fid)

			//// persist in dfs
			//if memSheet := getSheetCache().Get(fid); memSheet != nil {
			//	commitOneSheetWithCache(fid, memSheet)
			//	keys, evicted := getSheetCache().Put(fid)
			//	commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
			//} else {
			//	logger.Warnf("[fid:%d] Not in cache when deleting sheetws group", fid)
			//}
		}
	}
}

const (
	MsgAcquire	=	1
	MsgModify	=	2
	MsgRelease	=	3
)

type cellChanMarshal struct {
	MsgCode		int8				`json:"msgCode"`
	Content		string				`json:"content"`
	Info		json.RawMessage		`json:"info"`
	Uid			uint				`json:"uid"`
	Username	string				`json:"username"`
}

func sheetMessage2cellChanMarshalRaw(before *sheetMessage, uid uint, username string) (
	row int, col int, cellMsg *cellChanMarshal, err error) {
	after := cellChanMarshal{}
	switch before.MsgType {
	case "acquire":
		body := sheetAcquireMessage{}
		if err := json.Unmarshal(before.Body, &body); err != nil {
			return 0, 0, nil, err
		} else {
			row, col = body.Row, body.Col
			after.MsgCode = MsgAcquire
		}
	case "modify":
		body := sheetModifyMessage{}
		if err := json.Unmarshal(before.Body, &body); err != nil {
			return 0, 0, nil, err
		} else {
			row, col = body.Row, body.Col
			after.MsgCode = MsgModify
			after.Content = body.Content
			after.Info = body.Info
		}
	case "release":
		body := sheetReleaseMessage{}
		if err := json.Unmarshal(before.Body, &body); err != nil {
			return 0, 0, nil, err
		} else {
			row, col = body.Row, body.Col
			after.MsgCode = MsgRelease
		}
	default:
		return 0, 0, nil, errors.New("Unknown MsgType!")
	}

	after.Uid = uid
	after.Username = username
	return row, col, &after, nil
}

func SheetOnMessage(wss *wsWrap.WSServer, uid uint, username string, fid uint, body []byte) {
	if v, ok := sheetWSMeta.Load(fid); !ok {
		return
	} else {
		meta := v.(*sheetWSMetaEntry)
		// RLock here, cooperating with meta.wait() and meta.leave()
		meta.gLock.RLock()
		defer meta.gLock.RUnlock()		// TODO(flag): waiting

		sheetMsg := sheetMessage{}
		if err := json.Unmarshal(body, &sheetMsg); err != nil {
			return
		}

		if row, col, cellMsg, err := sheetMessage2cellChanMarshalRaw(&sheetMsg, uid, username); err != nil {
			logger.Errorf("[msgType(%s)\tfid(%d)\tuid(%d)] onMessage: Wrong format of sheet message <%v>",
				sheetMsg.MsgType, fid, uid, err)
		} else {
			key := cellKey{Row: row, Col: col}
			actual, loaded := meta.cellMap.LoadOrStore(key, &sheetWSCellMetaEntry{
				cellKey: key,
				workerSema: make(chan int, 1),
			})

			// wg.Add here, cooperating with meta.wait() and meta.leave()
			meta.cellWg.Add(1)

			cellMeta := actual.(*sheetWSCellMetaEntry)
			if !loaded {
				cellMeta.workerSema <- 1
			}
			if err := handleCellPool.Invoke(handleCellArgs{
				wss: wss,
				meta: meta,
				cellMeta: cellMeta,
				cellMsg: cellMsg,
			}); err != nil {
				meta.cellWg.Done()		// ignore message overflow simply
				logger.Debugf("handleCell Pool", err)
			}
		}
	}
}

// daemon goroutines: handleCell, handleNotify

func broadcast(meta *sheetWSMetaEntry, rawNtf []byte) {
	meta.userMap.Range(func(k interface{}, v interface{}) bool {
		userMeta := v.(*sheetWSUserMetaEntry)
		userMeta.notifyChan <- rawNtf
		return true
	})
}

func sheetModifyBroadcast(meta *sheetWSMetaEntry, cellMeta *sheetWSCellMetaEntry, cellMsg *cellChanMarshal) {
	bodyRaw, _ := json.Marshal(sheetModifyNotify{
		Row: cellMeta.cellKey.Row,
		Col: cellMeta.cellKey.Col,
		Content: cellMsg.Content,
		Info: cellMsg.Info,
		Username: cellMsg.Username,
	})

	rawNtf, _ := json.Marshal(sheetNotify{
		MsgType: "modify",
		Body: bodyRaw,
	})

	broadcast(meta, rawNtf)
}

func sheetPrepareBroadcast(msgType string, meta *sheetWSMetaEntry, cellMeta *sheetWSCellMetaEntry,
	cellMsg *cellChanMarshal) {
	bodyRaw, _ := json.Marshal(sheetPrepareNotify{
		Row: cellMeta.cellKey.Row,
		Col: cellMeta.cellKey.Col,
		Username: cellMsg.Username,
	})

	rawNtf, _ := json.Marshal(sheetNotify{
		MsgType: msgType,
		Body: bodyRaw,
	})

	broadcast(meta, rawNtf)
}

func doSheetAcquire(meta *sheetWSMetaEntry, cellMeta *sheetWSCellMetaEntry, cellMsg *cellChanMarshal) {
	if success := cellMeta.cellLock.tryLock(uint64(cellMsg.Uid)); success {
		sheetPrepareBroadcast("acquire", meta, cellMeta, cellMsg)
	} else {
		logger.Warnf("[fid(%d)\trow(%d)\tcol(%d)\tuid(%d)\tusername(%s)] " +
			"handleCell: acquiring a cell locked by others",
			meta.fid, cellMeta.cellKey.Row, cellMeta.cellKey.Col, cellMsg.Uid, cellMsg.Username)
	}
}

func doSheetRelease(meta *sheetWSMetaEntry, cellMeta *sheetWSCellMetaEntry, cellMsg *cellChanMarshal) {
	if success := cellMeta.cellLock.unLock(uint64(cellMsg.Uid)); success {
		sheetPrepareBroadcast("release", meta, cellMeta, cellMsg)
	} else {
		logger.Warnf("[fid(%d)\trow(%d)\tcol(%d)\tuid(%d)\tusername(%s)] " +
			"handleCell: releasing a cell locked by others",
			meta.fid, cellMeta.cellKey.Row, cellMeta.cellKey.Col, cellMsg.Uid, cellMsg.Username)
	}
}

func doSheetModify(meta *sheetWSMetaEntry, cellMeta *sheetWSCellMetaEntry, cellMsg *cellChanMarshal) {
	if success := cellMeta.cellLock.tryLock(uint64(cellMsg.Uid)); success {
		fid := meta.fid
		curCid := uint(sheetGetCheckPointNum(fid))
		inCache := true
		memSheet := getSheetCache().Get(fid)
		// not in sheetCache, load from log and do eviction if needed
		if memSheet == nil {
			if memSheet, inCache = recoverSheetFromLog(fid); memSheet == nil {
				panic("recoverSheetFromLog fails")
			}
		}

		row, col := cellMeta.cellKey.Row, cellMeta.cellKey.Col
		if memSheet.Get(row, col) != cellMsg.Content {
			// log
			lid := curCid + 1
			appendOneSheetLog(fid, lid, &gdocFS.SheetLogPickle{
				Lid: lid,
				Timestamp: time.Now(),
				Row: row,
				Col: col,
				Old: memSheet.Get(row, col),
				New: cellMsg.Content,
				Uid: cellMsg.Uid,
				Username: cellMsg.Username,
			})

			// cache
			memSheet.Set(row, col, cellMsg.Content)
			if !inCache {
				commitOneSheetWithCache(fid, memSheet)
			}

			sheetModifyBroadcast(meta, cellMeta, cellMsg)
		}

		if inCache {
			keys, evicted := getSheetCache().Put(fid)
			commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
		}


	} else {
		logger.Warnf("[fid(%d)\trow(%d)\tcol(%d)\tuid(%d)\tusername(%s)] " +
			"handleCell: modifying a cell locked by others",
			meta.fid, cellMeta.cellKey.Row, cellMeta.cellKey.Col, cellMsg.Uid, cellMsg.Username)
	}
}

type handleCellArgs struct {
	wss			*wsWrap.WSServer
	meta		*sheetWSMetaEntry
	cellMeta	*sheetWSCellMetaEntry
	cellMsg		*cellChanMarshal
}

func handleCellPoolTask(args interface{}) {
	cellArg := args.(handleCellArgs)
	handleCell(cellArg.wss, cellArg.meta, cellArg.cellMeta, cellArg.cellMsg)
}

func handleCell(wss *wsWrap.WSServer, meta *sheetWSMetaEntry, cellMeta *sheetWSCellMetaEntry,
	cellMsg *cellChanMarshal) {
	<- cellMeta.workerSema
	// wg.Done here, cooperating with meta.wait() and meta.leave()
	defer func () {
		cellMeta.workerSema <- 1
		meta.cellWg.Done()
	}()		// TODO(flag): waiting

	row, col := cellMeta.cellKey.Row, cellMeta.cellKey.Col

	logger.Debugf("[fid(%d)\trow(%d)\tcol(%d)] handleCell: run goroutine", meta.fid, row, col)

	switch cellMsg.MsgCode {
	case MsgAcquire:
		doSheetAcquire(meta, cellMeta, cellMsg)
	case MsgModify:
		doSheetModify(meta, cellMeta, cellMsg)
	case MsgRelease:
		doSheetRelease(meta, cellMeta, cellMsg)
	default:
		logger.Errorf("![fid(%d)\trow(%d)\tcol(%d)] handleCell: Unknown MsgType!",
			meta.fid, row, col)
	}
	return
}

func handleNotify(wss *wsWrap.WSServer, meta *sheetWSMetaEntry, userMeta *sheetWSUserMetaEntry) {
	logger.Debugf("[fid(%d)\tuid(%d)\tusername(%s)] handleNotify: run goroutine",
		meta.fid, userMeta.uid, userMeta.username)

	notifyChan, stopChan := &userMeta.notifyChan, &userMeta.stopChan
	for {
		select {
		case rawMsg := <- *notifyChan:
			id := utils.GenID("sheet", userMeta.uid, userMeta.username, meta.fid)
			wss.Send(id, rawMsg)
		case <- *stopChan:
			logger.Debugf("[fid(%d)\tuid(%d)\tusername(%s)] handleNotify: quit goroutine",
				meta.fid, userMeta.uid, userMeta.username)
			return
		}
	}
}

//
type cellLock struct {
	owner	uint64
}

func (cl *cellLock) tryLock(uid uint64) (success bool) {
	return atomic.LoadUint64(&cl.owner) == uid || atomic.CompareAndSwapUint64(&cl.owner, 0, uid)
}

func (cl *cellLock) unLock(uid uint64) (success bool) {
	return atomic.CompareAndSwapUint64(&cl.owner, uid, 0)
}


type cellLockNotify struct {
	Row			int
	Col			int
	Username	string
}

func dumpLocksOnCell(meta *sheetWSMetaEntry) (cellLocks []cellLockNotify) {
	meta.cellMap.Range(func(k, v interface{}) bool {
		key := k.(cellKey)
		cellMeta := v.(*sheetWSCellMetaEntry)
		ownerUid := cellMeta.cellLock.owner
		if ownerUid != 0 {
			if u, ok := meta.userMap.Load(uint(ownerUid)); ok {
				user := u.(*sheetWSUserMetaEntry)
				cellLocks = append(cellLocks, cellLockNotify{
					Row: key.Row,
					Col: key.Col,
					Username: user.username,
				})
			} else {
				logger.Warnf("[uid(%d)\trow(%d)\tcol(%d)] Cannot find user by lock owner (just leave)",
					ownerUid, key.Row, key.Col)
			}
		}
		return true
	})

	return cellLocks
}























