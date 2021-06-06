package main

import (
	"backend/server"
	"github.com/kataras/iris/v12"
	"syscall"
)

func main() {
	app := server.NewApp()

	_ = syscall.Dup2(1, 2)

	if err := app.Run(iris.Addr("0.0.0.0:8080"), iris.WithoutServerError(iris.ErrServerClosed)); err != nil {
		panic("Failed to Start Server!")
	}
}