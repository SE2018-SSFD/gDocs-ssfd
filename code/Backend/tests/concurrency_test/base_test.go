package concurrency_test

import (
	"backend/utils"
	"backend/utils/logger"
	"bytes"
	"encoding/json"
	"github.com/kataras/golog"
	"github.com/sacOO7/gowebsocket"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	wsApi = "sheetws"
)

var (
	hosts = []string{
		"192.168.1.107:10086",
		"192.168.1.107:10087",
		"192.168.1.107:10088",
	}

	loginParams = []utils.LoginParams{
		{"test", "test"},
		{"test1", "test1"},
		{"test2", "test2"},
	}
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	logger.SetLogger(golog.New())
	logger.SetLevel("Info")

	exitCode := m.Run()
	os.Exit(exitCode)
}

func randomHost() string {
	idx := rand.Int() % 3
	return hosts[idx]
}

func randomHostHttp() string {
	idx := rand.Int() % 3
	return "http://" + hosts[idx]
}

func get(addr string, api string, params string, respBody interface{}) (err error) {
	url := addr + "/" + api + "?" + params

	logger.Debugf("[%s] Send Get: %s", url, params)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	respBodyRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respBodyRaw = TransEscape(respBodyRaw)

	if respBody != nil {
		err = json.Unmarshal(respBodyRaw, respBody)
		if err != nil {
			return err
		}
	}

	logger.Debugf("[%s] Get Json Response: %v", url, respBody)

	return nil
}

func getRaw(addr string, api string, params string) (raw []byte, err error) {
	url := addr + "/" + api + "?" + params

	logger.Debugf("[%s] Send Get: %s", url, params)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	raw, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	raw = TransEscape(raw)

	logger.Debugf("[%s] Get Raw: %s", url, raw)
	return raw, err
}

func post(addr string, api string, reqBody interface{}, respBody interface{}) (err error) {
	url := addr + "/" + api
	reqBodyRaw, _ := json.Marshal(reqBody)

	logger.Debugf("[%s] Send Post: %s", url, reqBodyRaw)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyRaw))
	if err != nil {
		return err
	}

	respBodyRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respBodyRaw = TransEscape(respBodyRaw)

	logger.Debugf("[%s] Get Post Raw: %s", url, respBodyRaw)

	if respBody != nil {
		err = json.Unmarshal(respBodyRaw, respBody)
		if err != nil {
			return err
		}
	}

	logger.Debugf("[%s] Get Post Json Response: %v", url, respBody)

	return nil
}

func getPostRet(addr string, api string, reqBody interface{}, respBody interface{}) (err error) {
	bean := ResponseBean{}
	err = post(addr, api, reqBody, &bean)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bean.Data, respBody)
	if err != nil {
		return err
	}

	return nil
}

func getPostRaw(addr string, api string, reqBody interface{}) (raw []byte, err error) {
	bean := ResponseBean{}
	err = post(addr, api, reqBody, &bean)
	if err != nil {
		return nil, err
	}

	return bean.Data, nil
}

func getWSAddr(token string, fid uint) string {
	bean := ResponseBean{}
	addr := randomHost()
	err := get("http://" + addr, wsApi, "fid="+strconv.Itoa(int(fid))+"&token="+token+"&query=1", &bean)
	if err != nil {
		return ""
	}
	if !bean.Success {
		str := strings.Trim(string(bean.Data), "\"")
		logger.Debugf("[%q] Get WS address", str)
		return str
	} else {
		str := "ws://"+addr+"/"+wsApi+"?fid="+strconv.Itoa(int(fid))+"&token="+token
		logger.Debugf("[%q] Get WS address", str)
		return str
	}
}

func login(t *testing.T, params utils.LoginParams) (token string) {
	loginRet := LoginRet{}
	err := getPostRet(randomHostHttp(), "login", params, &loginRet)
	if assert.NoError(t, err) {
		return loginRet.Token
	} else {
		t.Fatal(err)
		return ""
	}
}

func TransEscape(data []byte) []byte {
	data = bytes.Replace(data, []byte("\\u0026"), []byte("&"), -1)
	data = bytes.Replace(data, []byte("\\u003c"), []byte("<"), -1)
	data = bytes.Replace(data, []byte("\\u003e"), []byte(">"), -1)
	return data
}

type onAcquireFunc func(msg sheetPrepareNotify)
type onModifyFunc func(msg sheetModifyNotify)
type onReleaseFunc func(msg sheetPrepareNotify)
type onConnFunc func(msg sheetOnConnNotify)

type myWS struct {
	ws	gowebsocket.Socket
}

type sheetMessage struct {
	MsgType		string				`json:"msgType"`	// acquire, modify, release, onConn
	Body		json.RawMessage		`json:"body"`
}

// client -> server
type sheetAcquireMessage struct {
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
}

type sheetModifyMessage struct {
	Row			int 			`json:"row"`
	Col			int				`json:"col"`
	Content		string			`json:"content"`
	Info		json.RawMessage	`json:"info"`
}

type sheetReleaseMessage struct {
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
}

// server -> client
type sheetPrepareNotify struct {
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
	Username	string		`json:"username"`
}

type sheetModifyNotify struct {
	Row			int 			`json:"row"`
	Col			int				`json:"col"`
	Content		string			`json:"content"`
	Info		json.RawMessage	`json:"info"`
	Username	string			`json:"username"`
}

type cellLockNotify struct {
	Row			int
	Col			int
	Username	string
}

type sheetOnConnNotify struct {
	Name			string				`json:"name"`
	Columns			int					`json:"columns"`
	Content			[]string			`json:"content"`
	CellLocks		[]cellLockNotify	`json:"cellLocks"`
}

func NewWebSocket(t *testing.T, addr string,
	func1 onAcquireFunc, func2 onModifyFunc, func3 onReleaseFunc, func4 onConnFunc) (ws *myWS) {

	ws = &myWS{
		ws: gowebsocket.New(addr),
	}

	ws.ws.ConnectionOptions = gowebsocket.ConnectionOptions{
		UseSSL: false,
		UseCompression: true,
	}

	t.Log(ws.ws.RequestHeader.Clone())

	ws.ws.OnTextMessage = func(content string, socket gowebsocket.Socket) {
		data := []byte(content)
		bean := sheetMessage{}
		if err := json.Unmarshal(data, &bean); err != nil {
			t.Error(err)
		} else {
			switch bean.MsgType {
			case "acquire":
				msg := sheetPrepareNotify{}
				err = json.Unmarshal(bean.Body, &msg)
				if err != nil {
					t.Error(err)
				} else {
					func1(msg)
				}
			case "modify":
				msg := sheetModifyNotify{}
				err = json.Unmarshal(bean.Body, &msg)
				if err != nil {
					t.Error(err)
				} else {
					func2(msg)
				}
			case "release":
				msg := sheetPrepareNotify{}
				err = json.Unmarshal(bean.Body, &msg)
				if err != nil {
					t.Error(err)
				} else {
					func3(msg)
				}
			case "onConn":
				msg := sheetOnConnNotify{}
				err = json.Unmarshal(bean.Body, &msg)
				if err != nil {
					t.Error(err)
				} else {
					func4(msg)
				}
			default:
				t.Errorf("[%s] No matched msgType", bean.MsgType)
			}
		}
	}
	ws.ws.OnConnectError = func(err error, socket gowebsocket.Socket) {
		t.Error(err)
	}
	ws.ws.Connect()
	return ws
}

func (ws *myWS) SendJson(msgType string, msg interface{}) (err error) {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	sheetMsg := sheetMessage{
		MsgType: msgType,
		Body: raw,
	}

	raw, err = json.Marshal(sheetMsg)
	if err != nil {
		return err
	}

	ws.ws.SendBinary(raw)
	return nil
}


type ResponseBean struct {
	Success		bool			`json:"success"`
	Msg			int				`json:"msg"`
	Data		json.RawMessage	`json:"data"`
}

type LoginRet struct {
	Info	json.RawMessage		`json:"info"`
	Token	string				`json:"token"`
}


type GetSheetRet struct {
	Fid						uint		`json:"fid"`
	IsDeleted				bool		`json:"isDeleted"`
	Name					string		`json:"name"`
	CheckPointNum			int			`json:"checkpoint_num"`
	Columns					int			`json:"columns"`
	Owner					string		`json:"owner"`

	CreatedAt 				time.Time	`json:"CreatedAt"`
	UpdatedAt 				time.Time	`json:"UpdatedAt"`

	Content					[]string	`json:"content"`
}