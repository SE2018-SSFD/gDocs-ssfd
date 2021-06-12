package server

import (
	"backend/repository"
	"backend/router"
	"backend/utils/config"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

func NewApp() *iris.Application {
	app := iris.New()
	app.Logger().SetLevel("debug")

	app.Use(recover.New())
	app.Use(logger.New())

	router.SetRouter(app)
	repository.InitDBConn()
	config.LoadConfig("")

	return app
}
