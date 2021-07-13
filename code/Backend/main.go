package main

import (
	"backend/server"
	"github.com/kataras/iris/v12"
)

func main() {
	app := server.NewApp()

	if err := app.Run(iris.Addr("0.0.0.0:8080"), iris.WithoutServerError(iris.ErrServerClosed)); err != nil {
		panic("Failed to Start Server!")
	}
}