package system_test

import (
	"backend/server"
	"github.com/iris-contrib/httpexpect/v2"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/kataras/iris/v12/httptest"
	"github.com/kataras/iris/v12/i18n"
	"net/http"
	"os"
	"testing"
)

var (
	app			*iris.Application
	testToken1	string
	testToken2	string
	testToken3	string
)

func TestMain(m *testing.M) {
	app = server.NewApp()
	testToken1 = "0000000000000000000000000000000000000000000000000000000000000000"
	testToken2 = "1111111111111111111111111111111111111111111111111111111111111111"
	testToken3 = "2222222222222222222222222222222222222222222222222222222222222222"
	exitCode := m.Run()
	os.Exit(exitCode)
}

func post(t *testing.T, path string, Object interface{}, StatusCode int, success bool, Msg int, data interface{}) *httpexpect.Object {
	e := getHttpExpect(t)
	var testMap map[string]interface{}

	if data != nil {
		testMap = map[string]interface{} {
			"success": success,
			"msg": Msg,
			"data": data,
		}
	} else {
		testMap = map[string]interface{} {
			"success": success,
			"msg": Msg,
		}
	}

	return e.POST(path).WithJSON(Object).
		Expect().Status(StatusCode).
		JSON().Object().ContainsMap(testMap)
}


func getHttpExpect(t *testing.T) *httpexpect.Expect {
	return httptest.New(t, app, httptest.Configuration{Debug: true, URL: "http://localhost:8080"})
}

func Do(c **context.Context, w http.ResponseWriter, r *http.Request, handler iris.Handler, irisConfigurators ...iris.Configurator) {
	app := new(iris.Application)
	app.I18n = i18n.New()
	app.Configure(iris.WithConfiguration(iris.DefaultConfiguration()), iris.WithLogLevel("disable"))
	app.Configure(irisConfigurators...)

	app.HTTPErrorHandler = router.NewDefaultHandler(app.ConfigurationReadOnly(), app.Logger())
	app.ContextPool = context.New(func() interface{} {
		return context.NewContext(app)
	})

	ctx := app.ContextPool.Acquire(w, r)
	*c = ctx
	handler(ctx)
	app.ContextPool.Release(ctx)
}

