package controller

import (
	"backend/lib/wsWrap"
	"backend/service"
	"backend/utils"
	"github.com/kataras/iris/v12/context"
	"github.com/panjf2000/ants/v2"
	"strconv"
)

var wss *wsWrap.WSServer

var onMessagePool *ants.PoolWithFunc

const (
	onMessagePoolSize = 2000
)

func init() {
	wss = wsWrap.NewWSServer(onConn, onDisConn, onMessage)
	var err error
	onMessagePool, err = ants.NewPoolWithFunc(onMessagePoolSize, onMessagePoolTask, ants.WithNonblocking(true))
	if err != nil {
		panic(err)
	}
}

func onConn(id string) {
	ns, uid, username, fid := utils.ParseID(id)
	switch ns {
	case "sheet":
		service.SheetOnConn(wss, uid, username, fid)
	}
}

func onDisConn(id string) {
	ns, uid, username, fid := utils.ParseID(id)
	switch ns {
	case "sheet":
		service.SheetOnDisConn(wss, uid, username, fid)
	}
}

type onMessageArgs struct {
	uid			uint
	username	string
	fid			uint
	body		[]byte
}

func onMessagePoolTask(args interface{}) {
	msgArg := args.(*onMessageArgs)
	service.SheetOnMessage(wss, msgArg.uid, msgArg.username, msgArg.fid, msgArg.body)
}

func onMessage(id string, body []byte) {
	ns, uid, username, fid := utils.ParseID(id)
	switch ns {
	case "sheet":
		_ = onMessagePool.Invoke(&onMessageArgs{
			uid: uid,
			username: username,
			fid: fid,
			body: body,
		})
	}
}

func SheetUpgradeHandler() context.Handler {
	idGen := func(ctx *context.Context) string {
		ns := ctx.Params().Get("ns")
		uid, _ := ctx.Params().GetUint("uid")
		username := ctx.Params().Get("username")
		fid, _ := ctx.Params().GetUint("fid")
		return utils.GenID(ns, uid, username, fid)
	}
	return wss.Handler(idGen)
}

func SheetBeforeUpgradeHandler() context.Handler {
	return func(ctx *context.Context) {
		fid := uint(ctx.URLParamUint64("fid"))
		token := ctx.URLParam("token")
		query := ctx.URLParamUint64("query")
		if success, msg, user, addr := service.SheetOnConnEstablished(token, fid); !success {
			if addr != "" {
				utils.SendResponse(ctx, success, msg, "ws://" + addr +
					"/sheetws?" + "token=" + token + "&" + "fid=" + strconv.Itoa(int(fid)))
				ctx.StopExecution()
			} else {
				utils.SendResponse(ctx, success, msg, nil)
				ctx.StopExecution()
			}
		} else {
			if query != 0 {
				utils.SendResponse(ctx, true, utils.SheetWSDestination, nil)
				ctx.StopExecution()
			} else {
				ctx.Params().Save("ns", "sheet", false)
				ctx.Params().Save("uid", user.Uid, false)
				ctx.Params().Save("username", user.Username, false)
				ctx.Params().Save("fid", fid, false)
				ctx.Next()
			}
		}
	}
}