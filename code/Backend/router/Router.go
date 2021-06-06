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
	root.Handle("POST", "/modifysheet", controller.ModifySheet)
	root.Handle("POST", "/deletesheet", controller.DeleteSheet)
	root.Handle("POST", "/commitsheet", controller.CommitSheet)
	root.Handle("POST", "/getchunk", controller.GetChunk)

	return
}