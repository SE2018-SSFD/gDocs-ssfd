package wsWrap

import (
	"backend/utils/logger"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"
	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gorilla"
	"net/http"
	"sync"
	"time"
)

type WSServer struct {
	server		*neffos.Server
	connMap		sync.Map
}

type IDGenerator = websocket.IDGenerator

type OnMessageCallback func(id string, body []byte)
type OnConnCallback func(id string)
type OnDisConnCallback func(id string)

func NewWSServer(onConnCallback OnConnCallback, onDisConnCallback OnDisConnCallback,
	onMessageCallback OnMessageCallback) *WSServer {
	wss := WSServer{
		server: websocket.New(
			gorilla.Upgrader(gorillaWs.Upgrader{CheckOrigin: func(*http.Request) bool{ return true }}),
			websocket.Events{
				websocket.OnNativeMessage: func(nsConn *neffos.NSConn, msg neffos.Message) error {
					id := nsConn.Conn.ID()

					go onMessageCallback(id, msg.Body)
					return nil
				},
			},
		),
	}

	wss.server.OnConnect = func(conn *neffos.Conn) error {
		wss.connMap.Store(conn.ID(), conn)
		go onConnCallback(conn.ID())
		return nil
	}

	wss.server.OnDisconnect = func(conn *neffos.Conn) {
		wss.connMap.Delete(conn.ID())
		go onDisConnCallback(conn.ID())
	}

	return &wss
}


func (wss *WSServer) Handler(idGen IDGenerator) iris.Handler {
	return websocket.Handler(wss.server, idGen)
}

func (wss *WSServer) Send(id string, content []byte) {
	if v, ok := wss.connMap.Load(id); ok {
		conn := v.(*websocket.Conn)
		for !conn.IsClosed() {
			// TODO: races may exist?
			if err := conn.Socket().WriteText(content, time.Second * 5); err == nil {
				break
			}
		}
	} else {
		logger.Errorf("[%s] ID does not exist")
	}
}