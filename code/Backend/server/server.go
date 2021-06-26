package server

import (
	"backend/lib/cluster"
	"backend/lib/zkWrap"
	"backend/repository"
	"backend/router"
	"backend/utils/config"
	loggerWrap "backend/utils/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

func NewApp() *iris.Application {
	app := iris.New()
	app.Logger().SetLevel("debug")

	app.Use(recover.New())
	app.Use(logger.New())

	loggerWrap.SetLogger(app.Logger())

	config.LoadConfig()

	router.SetRouter(app)

	repository.InitDBConn()

	if !config.Get().WriteThrough {
		if err := zkWrap.Chroot(config.Get().ZKRoot); err != nil {
			panic(err)
		}
		cluster.RegisterNodes(config.Get().Addr, int(config.Get().MaxSheetCache/config.Get().UnitCache))
	}

	return app
}
