package router

import (
	"backend/controller"
	"backend/middleware"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/pprof"
)

func SetRouter(app *iris.Application) {
	root := app.Party("/", middleware.CrsAuth()).AllowMethods(iris.MethodOptions)

	root.Handle("POST", "/login", controller.Login)
	root.Handle("POST", "/register", controller.Register)
	root.Handle("POST", "/getuser", controller.GetUser)
	root.Handle("POST", "/modifyuser", controller.ModifyUser)
	root.Handle("POST", "/modifyuserauth", controller.ModifyUserAuth)
	root.Handle("POST", "/newsheet", controller.NewSheet)
	root.Handle("POST", "/getsheet", controller.GetSheet)
	root.Handle("POST", "/deletesheet", controller.DeleteSheet)
	root.Handle("POST", "/recoversheet", controller.RecoverSheet)
	root.Handle("POST", "/commitsheet", controller.CommitSheet)
	root.Handle("GET", "/getchunk", controller.GetChunk)
	root.Handle("POST", "/uploadchunk", controller.UploadChunk)
	root.Handle("POST", "/getallchunks", controller.GetAllChunks)
	root.Handle("POST", "/getsheetchkp", controller.GetSheetCheckPoint)
	root.Handle("POST", "/getsheetlog", controller.GetSheetLog)
	root.Handle("POST", "/rollbacksheet", controller.RollbackSheet)

	root.Handle("GET", "/sheetws",
		controller.SheetBeforeUpgradeHandler(), controller.SheetUpgradeHandler())

	// for debug
	p := pprof.New()
	root.Handle("GET", "/pprof", p)
	root.Handle("GET", "/pprof/{action:path}", p)

	return
}