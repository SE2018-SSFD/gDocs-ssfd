package sheetWS

import (
	"backend/dao"
	"backend/model"
	"backend/service"
	"backend/utils"
	"encoding/json"
	"errors"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/websocket"
	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gorilla"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var sheetWSServer *neffos.Server

var sheetGroup sync.Map

type sheetGroupEntry struct {
	fid			uint
	userMap		sync.Map
	userN		int
}

type sheetUserEntry struct {
	uid			uint
	username	string
	conn		*websocket.Conn
}

type sheetMessage struct {
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
	Content		string		`json:"content"`
}

func init() {
	sheetWSServer = websocket.New(gorilla.Upgrader(gorillaWs.Upgrader{CheckOrigin: func(*http.Request) bool{return true}}),
		websocket.Events{
			websocket.OnNativeMessage: func(nsConn *websocket.NSConn, msg websocket.Message) error {
				uid, _, fid := parseID(nsConn.Conn.ID())
				log.Println("onMessage:", uid, fid)
				if v, ok := sheetGroup.Load(fid); !ok {
					return errors.New("group does not exists")
				} else {
					group := v.(*sheetGroupEntry)
					sheetMsg := sheetMessage{}
					if err := json.Unmarshal(msg.Body, &sheetMsg); err != nil {
						return errors.New("wrong format of sheet websocket message")
					}
					// TODO: lock
					service.ModifySheetCache(fid, sheetMsg.Row, sheetMsg.Col, sheetMsg.Content)
					// TODO: do log asynchronously
					group.userMap.Range(func (k interface{}, v interface{}) bool {
						curUid := k.(uint)
						curUser := v.(*sheetUserEntry)
						if curUid != uid {
							// TODO: implement queued write (is concurrent write async-safe?)
							go curUser.conn.Write(msg)
						}
						return true
					})
					return nil
				}
			}})

	sheetWSServer.OnConnect = func(conn *websocket.Conn) error {
		log.Printf("[%s] Connected to server!", conn.ID())
		uid, username, fid := parseID(conn.ID())
		if v, ok := sheetGroup.Load(fid); !ok {
			return errors.New("group does not exists")
		} else {
			user := sheetUserEntry{
				uid:      uid,
				username: username,
				conn: conn,
			}
			group := v.(*sheetGroupEntry)
			group.userN += 1
			group.userMap.Store(uid, &user)
			return nil
		}
	}

	sheetWSServer.OnDisconnect = func(conn *websocket.Conn) {
		log.Printf("[%s] Disconnected from server", conn.ID())
		uid, _, fid := parseID(conn.ID())
		group := safeLoad(&sheetGroup, fid).(*sheetGroupEntry)
		group.userMap.Delete(uid)
		group.userN -= 1
		if group.userN == 0 {
			log.Printf("[%d] delete group", fid)
			sheetGroup.Delete(fid)
		}
	}
}

func UpgradeHandler() context.Handler {
	idGen := func(ctx *context.Context) string {
		return ctx.Params().Get("uid") + "#" + ctx.Params().Get("username") + "#" + ctx.Params().Get("fid")
	}
	return func(ctx *context.Context) {
		log.Println(ctx.Params().Get("uid") + "#" + ctx.Params().Get("username") + "#" + ctx.Params().Get("fid"))
		websocket.Handler(sheetWSServer, idGen)(ctx)
	}
}

func BeforeUpgradeHandler() context.Handler {
	return func(ctx *context.Context) {
		fid := uint(ctx.URLParamUint64("fid"))
		token := ctx.URLParam("token")
		if success, msg, user := checkTokenAndSheet(token, fid); !success {
			utils.SendResponse(ctx, success, msg, nil)
			ctx.StopExecution()
		} else {
			sheetGroup.LoadOrStore(fid, &sheetGroupEntry{
				fid: fid,
				userN: 0,
			})

			ctx.Params().Save("uid", strconv.FormatUint(uint64(user.Uid), 10), false)
			ctx.Params().Save("username", user.Username, false)
			ctx.Params().Save("fid", strconv.FormatUint(uint64(fid), 10), false)
			ctx.Next()
		}
	}
}

func checkTokenAndSheet(token string, fid uint) (bool, int, *model.User) {
	if token == "" || fid == 0 {
		return false, utils.InvalidFormat, nil
	}

	uid := service.CheckToken(token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, fid) {
			return false, utils.SheetNoPermission, nil
		} else {
			user := dao.GetUserByUid(uid)
			return true, 0, &user
		}
	} else {
		return false, utils.InvalidToken, nil
	}
}

func parseID(id string) (uint, string, uint) {
	split := strings.SplitN(id, "#", 3)
	uid, _ := strconv.ParseUint(split[0], 10, 64)
	username := split[1]
	fid, _ := strconv.ParseUint(split[2], 10, 64)
	return uint(uid), username, uint(fid)
}

func safeLoad(m *sync.Map, key interface{}) interface{} {
	if v, ok := m.Load(key); !ok {
		panic("key does not exist")
	} else {
		return v
	}
}