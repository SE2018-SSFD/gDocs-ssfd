package router

import (
	"backend/controller"
	"backend/middleware"
	"github.com/kataras/iris/v12"
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
	root.Handle("POST", "/getchunk", controller.GetChunk)
	root.Handle("POST", "/getsheetchkp", controller.GetSheetCheckPoint)
	root.Handle("POST", "/getsheetlog", controller.GetSheetLog)

	root.Handle("GET", "/sheetws",
		controller.SheetBeforeUpgradeHandler(), controller.SheetUpgradeHandler())

	return
}