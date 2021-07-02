package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/lib/cluster"
	"backend/lib/gdocFS"
	"backend/lib/wsWrap"
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"backend/utils/logger"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
)

var sheetGroup sync.Map

type cellKey struct {
	Row		int
	Col		int
}

type cellLock struct {
	owner	uint64
}

type sheetGroupEntry struct {
	fid			uint
	userMap		sync.Map
	userN		int
	lockMap		sync.Map
}

type sheetUserEntry struct {
	uid			uint
	username	string
}

type sheetMessage struct {
	MsgType		string				`json:"msgType"`	// acquire, modify, release
	Body		json.RawMessage		`json:"body"`
}

// client -> server
type sheetAcquireMessage struct {
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
}

type sheetModifyMessage struct {
	Row			int 			`json:"row"`
	Col			int				`json:"col"`
	Content		string			`json:"content"`
	Info		json.RawMessage	`json:"info"`
}

type sheetReleaseMessage struct {
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
}

// server -> client
type sheetPrepareNotify struct {
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
	Username	string		`json:"username"`
}

type sheetModifyNotify struct {
	Row			int 			`json:"row"`
	Col			int				`json:"col"`
	Content		string			`json:"content"`
	Info		json.RawMessage	`json:"info"`
	Username	string			`json:"username"`
}

func SheetOnConn(uid uint, username string, fid uint) {
	logger.Debugf("[%d %s %d] Connected to server!", uid, username, fid)
	if v, ok := sheetGroup.Load(fid); !ok {
		logger.Errorf("[fid:%d username:%s uid:%d] No group entry for sheetws!", fid, username, uid)
	} else {
		user := sheetUserEntry{
			uid:      uid,
			username: username,
		}
		group := v.(*sheetGroupEntry)
		if _, loaded := group.userMap.LoadOrStore(uid, &user); !loaded {
			group.userN += 1
		}
	}
}

func SheetOnDisConn(uid uint, username string, fid uint) {
	logger.Debugf("[%d %s %d] DisConnected from server!", uid, username, fid)

	if v, ok := sheetGroup.Load(fid); !ok {
		logger.Errorf("[fid:%d username:%s uid:%d] No group entry for sheetws!", fid, username, uid)
	} else {
		group := v.(*sheetGroupEntry)
		group.userMap.Delete(uid)
		group.userN -= 1
		if group.userN == 0 {
			logger.Debugf("[%d] delete group", fid)
			// delete group entry
			sheetGroup.Delete(fid)

			// persist in dfs
			if !config.Get().WriteThrough {
				if memSheet := getSheetCache().Get(fid); memSheet != nil {
					commitSheetsWithCache([]uint{fid}, []*cache.MemSheet{memSheet})
				} else {
					logger.Warnf("[fid:%d] Recovering from log", fid)
					sheet := dao.GetSheetByFid(fid)
					if memSheet = recoverSheetFromLog(&sheet); memSheet == nil {
						logger.Errorf("[fid:%d] Recover from log fails!", fid)
						return
					}
					commitSheetsWithCache([]uint{fid}, []*cache.MemSheet{memSheet})
				}
			}
		}
	}
}

func SheetOnMessage(wss *wsWrap.WSServer, uid uint, username string, fid uint, body []byte) {
	if v, ok := sheetGroup.Load(fid); !ok {
		return
	} else {
		group := v.(*sheetGroupEntry)
		sheetMsg := sheetMessage{}
		if err := json.Unmarshal(body, &sheetMsg); err != nil {
			return
		}

		logger.Debugf("[msgType:%s fid:%d uid:%d] onMessage", sheetMsg.MsgType, fid, uid)
		if config.Get().WriteThrough {
			handleSheetMessageWriteThrough(wss, fid, uid, username, sheetMsg, group)
		} else {
			handleSheetMessageWithCache(wss, fid, uid, username, sheetMsg, group)
		}

	}
}

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

			if !config.Get().WriteThrough {
				if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
					return false, utils.SheetWSRedirect, nil, addr
				}
			}

			user := dao.GetUserByUid(uid)

			if actual, loaded := sheetGroup.LoadOrStore(fid, &sheetGroupEntry{
				fid: fid,
				userN: 0,
			}); loaded {
				group := actual.(*sheetGroupEntry)
				if _, ok := group.userMap.Load(uid); ok {
					return false, utils.SheetDupConnection, nil, ""
				}
			}

			return true, 0, &user, ""
		}
	} else {
		return false, utils.InvalidToken, nil, ""
	}
}

func handleSheetMessageWriteThrough(wss *wsWrap.WSServer, fid uint, uid uint, username string,
	sheetMsg sheetMessage, group *sheetGroupEntry) {
	switch sheetMsg.MsgType {
	case "acquire":
		doSheetAcquire(wss, fid, uid, username, sheetMsg, group)
	case "modify":
		doSheetModifyWriteThrough(wss, fid, uid, username, sheetMsg, group)
	case "release":
		doSheetRelease(wss, fid, uid, username, sheetMsg, group)
	default:
		logger.Errorf("[fid:%d msgType:%s] Unknown type of sheet message!", fid, sheetMsg.MsgType)
	}
}

func handleSheetMessageWithCache(wss *wsWrap.WSServer, fid uint, uid uint, username string,
	sheetMsg sheetMessage, group *sheetGroupEntry) {
	switch sheetMsg.MsgType {
	case "acquire":
		doSheetAcquire(wss, fid, uid, username, sheetMsg, group)
	case "modify":
		doSheetModifyWithCache(wss, fid, uid, username, sheetMsg, group)
	case "release":
		doSheetRelease(wss, fid, uid, username, sheetMsg, group)
	default:
		logger.Errorf("[fid:%d msgType:%s] Unknown type of sheet message!", fid, sheetMsg.MsgType)
	}
}

func tryLockOnCell(group *sheetGroupEntry, uid uint, row int, col int) bool {
	if actual, loaded := group.lockMap.LoadOrStore(
		cellKey{Row: row, Col: col}, &cellLock{owner: uint64(uid)}); loaded {
		lock := actual.(*cellLock)
		return atomic.LoadUint64(&lock.owner) == uint64(uid) ||
			atomic.CompareAndSwapUint64(&lock.owner, 0, uint64(uid))
	} else {
		return true
	}
}

func unlockOnCell(group *sheetGroupEntry, uid uint, row int, col int) (success bool) {
	if v, ok := group.lockMap.Load(cellKey{Row: row, Col: col}); ok {
		lock := v.(*cellLock)
		success = atomic.CompareAndSwapUint64(&lock.owner, uint64(uid), 0)
		if !success {
			logger.Warnf("[uid:%d row:%d col:%d] Unlock the cell lock that not owned by the user", uid, row, col)
		}
		return success
	} else {
		logger.Warnf("[uid:%d row:%d col:%d] Cell lock does not exist", uid, row, col)
		return false
	}
}

func broadcast(wss *wsWrap.WSServer, group *sheetGroupEntry, uid uint, fid uint, content []byte) {
	group.userMap.Range(func (k interface{}, v interface{}) bool {
		curUid := k.(uint)
		curUser := v.(*sheetUserEntry)
		if curUid != uid {
			go wss.Send(utils.GenID("sheet", curUid, curUser.username, fid), content)
		}
		return true
	})
}

func doSheetAcquire(wss *wsWrap.WSServer, fid uint, uid uint, username string,
	sheetMsg sheetMessage, group *sheetGroupEntry) {
	msg := sheetAcquireMessage{}
	if err := json.Unmarshal(sheetMsg.Body, &msg); err != nil {
		logger.Errorf("[fid:%d msgType:%s] Wrong format of sheet message!", fid, sheetMsg.MsgType)
		return
	}
	if success := tryLockOnCell(group, uid, msg.Row, msg.Col); success {
		sheetPrepareBroadcast(wss, msg.Row, msg.Col, username, "acquire", fid, uid, group)
	}
}

func doSheetRelease(wss *wsWrap.WSServer, fid uint, uid uint, username string,
	sheetMsg sheetMessage, group *sheetGroupEntry) {
	msg := sheetReleaseMessage{}
	if err := json.Unmarshal(sheetMsg.Body, &msg); err != nil {
		logger.Errorf("[fid:%d msgType:%s] Wrong format of sheet message!", fid, sheetMsg.MsgType)
		return
	}
	if success := unlockOnCell(group, uid, msg.Row, msg.Col); success {
		sheetPrepareBroadcast(wss, msg.Row, msg.Col, username, "release", fid, uid, group)
	}
}

func doSheetModifyWriteThrough(wss *wsWrap.WSServer, fid uint, uid uint, username string,
	sheetMsg sheetMessage, group *sheetGroupEntry) {
	if msg := sheetModifyAuthenticateCell(fid, uid, username, sheetMsg, group); msg != nil {
		path := gdocFS.GetCheckPointPath("sheet", fid, 0)
		fileRaw, err := dao.FileGetAll(path)
		if err != nil {
			logger.Errorf("[%s] Checkpoint file does not exist!", path)
			return
		}

		filePickled, err := gdocFS.PickleSheetCheckPointFromContent(fileRaw)
		if err != nil {
			logger.Errorf("[%s] Checkpoint file cannot be pickled!", path)
			return
		}

		if msg.Row >= filePickled.Rows || msg.Col >= filePickled.Columns {
			tmpCells := cache.NewCellNetFromStringSlice(filePickled.Content, filePickled.Columns)
			tmpCells.Set(msg.Row, msg.Col, msg.Content)
			filePickled.Content = tmpCells.ToStringSlice()
			filePickled.Rows, filePickled.Columns = tmpCells.Shape()
		} else {
			filePickled.Content[msg.Row * filePickled.Columns + msg.Col] = msg.Content
		}

		filePickled.Timestamp = time.Now()
		fileRawByte, _ := json.Marshal(filePickled)
		fileRaw = string(fileRawByte)
		if err = dao.FileOverwriteAll(path, fileRaw); err != nil {
			logger.Errorf("[%s] Checkpoint file overwrite fails!", path)
			return
		}

		sheet := dao.GetSheetByFid(fid)
		sheet.Columns = filePickled.Columns
		dao.SetSheet(sheet)

		sheetModifyBroadcast(wss, fid, uid, username, group, msg)
	}
}

func doSheetModifyWithCache(wss *wsWrap.WSServer, fid uint, uid uint, username string,
	sheetMsg sheetMessage, group *sheetGroupEntry) {
	if msg := sheetModifyAuthenticateCell(fid, uid, username, sheetMsg, group); msg != nil {
		sheet := dao.GetSheetByFid(fid)

		memSheet := getSheetCache().Get(fid)
		// not in sheetCache, load from log and do eviction if needed
		if memSheet == nil {
			if memSheet = recoverSheetFromLog(&sheet); memSheet == nil {
				panic("recoverSheetFromLog fails")
			}
		}

		// log
		lid := uint(sheet.CheckPointNum) + 1
		row, col := msg.Row, msg.Col
		appendOneSheetLog(fid, lid, &gdocFS.SheetLogPickle{
			Lid: lid,
			Timestamp: time.Now(),
			Row: row,
			Col: col,
			Old: memSheet.Get(row, col),
			New: msg.Content,
			Uid: uid,
			Username: username,
		})

		// cache
		memSheet.Set(msg.Row, msg.Col, msg.Content)
		_, sheet.Columns = memSheet.Shape()
		dao.SetSheet(sheet)

		sheetModifyBroadcast(wss, fid, uid, username, group, msg)
	}
}

func sheetModifyAuthenticateCell(fid uint, uid uint, username string,
	sheetMsg sheetMessage, group *sheetGroupEntry) *sheetModifyMessage {
	msg := sheetModifyMessage{}
	if err := json.Unmarshal(sheetMsg.Body, &msg); err != nil {
		logger.Errorf("[fid:%d msgType:%s] Wrong format of sheet message!", fid, sheetMsg.MsgType)
		return nil
	}

	if success := tryLockOnCell(group, uid, msg.Row, msg.Col); !success {
		logger.Errorf("[fid:%d msgType:%s uid:%d username:%s] User modify a cell without lock!",
			fid, sheetMsg.MsgType, uid, username)
		return nil
	}

	return &msg
}

func sheetModifyBroadcast(wss *wsWrap.WSServer, fid uint, uid uint, username string,
	group *sheetGroupEntry, msg *sheetModifyMessage) {
	bodyRaw, _ := json.Marshal(sheetModifyNotify{
		Row: msg.Row,
		Col: msg.Col,
		Content: msg.Content,
		Info: msg.Info,
		Username: username,
	})

	toBroadcast, _ := json.Marshal(sheetMessage{
		MsgType: "modify",
		Body: bodyRaw,
	})

	broadcast(wss, group, uid, fid, toBroadcast)
}

func sheetPrepareBroadcast(wss *wsWrap.WSServer, row int, col int, username string, msgType string, fid uint, uid uint,
	group *sheetGroupEntry) {
	bodyRaw, _ := json.Marshal(sheetPrepareNotify{
		Row: row,
		Col: col,
		Username: username,
	})

	toBroadcast, _ := json.Marshal(sheetMessage{
		MsgType: msgType,
		Body: bodyRaw,
	})

	broadcast(wss, group, uid, fid, toBroadcast)
}