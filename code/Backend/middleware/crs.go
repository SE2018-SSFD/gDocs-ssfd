package middleware

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12/context"
)

func CrsAuth() context.Handler {
	return cors.AllowAll()
}