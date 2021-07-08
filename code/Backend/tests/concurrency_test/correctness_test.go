package concurrency_test

import (
	"backend/lib/cache"
	"backend/utils"
	"backend/utils/logger"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSingleFile(t *testing.T) {
	goTestWorkFlowInterrupt(t, 0)
}

//func TestMultiFile(t *testing.T) {
//	testRoomN := 30
//	wg := sync.WaitGroup{}
//	wg.Add(testRoomN)
//	for i := 0; i < testRoomN; i += 1 {
//		go func() {
//			goTestWorkFlowInterrupt(t, time.Second)
//			wg.Done()
//		}()
//		time.Sleep(time.Second)
//	}
//	wg.Wait()
//}

func goTestWorkFlowInterrupt(t *testing.T, pause time.Duration) {
	testWsNum := 20
	initRow, initCol := 10, 10
	tokenS, tokenR := login(t, loginParams[0]), login(t, loginParams[1])
	newSheetArg := utils.NewSheetParams{
		Token: tokenS,
		Name:  "test",
		InitRows: uint(initRow),
		InitColumns: uint(initCol),
	}
	raw, _ := getPostRaw(randomHostHttp(), "newsheet", newSheetArg)
	fidInt, err := strconv.Atoi(string(raw))
	fid := uint(fidInt)
	if assert.NoError(t, err) {
		getSheetArg := utils.GetSheetParams{
			Token: tokenR,
			Fid: fid,
		}
		getSheetRet := GetSheetRet{}
		err = getPostRet(randomHostHttp(), "getsheet", getSheetArg, &getSheetRet)
		if assert.NoError(t, err) {
			assert.EqualValues(t, fid, getSheetRet.Fid)
			assert.EqualValues(t, 0, getSheetRet.CheckPointNum)
			assert.Equal(t, "test", getSheetRet.Name)

			row, col := len(getSheetRet.Content) / getSheetRet.Columns, getSheetRet.Columns
			assert.Equal(t, len(getSheetRet.Content), row * col)

			ms := make([]*cache.MemSheet, testWsNum)
			for i := 0; i < testWsNum; i += 1 {
				ms[i] = cache.NewMemSheet(initRow, initCol)
			}

			var msSender, msReceiver *cache.MemSheet
			var mapS, mapR sync.Map

			type cellKey struct {
				Row int
				Col int
			}

			cnt := 0
			onAcquireS := func(msg sheetPrepareNotify) {}
			onModifyS := func(msg sheetModifyNotify) {
				logger.Info(cnt)
				cnt++
				logger.Debugf("[%+v] onModify", msg)
				msSender.Set(msg.Row, msg.Col, msg.Content)
				mapS.Store(cellKey{Row: msg.Row, Col: msg.Col}, msg.Content)
			}
			onReleaseS := func(msg sheetPrepareNotify) {}
			onConnS := func(msg sheetOnConnNotify) {
				logger.Debugf("[%+v] onConnection", msg)
				msSender = cache.NewMemSheetFromStringSlice(msg.Content, msg.Columns)
				for i := 0; i < len(msg.Content); i += 1 {
					mapS.Store(cellKey{Row: i/msg.Columns, Col: i%msg.Columns}, msg.Content[i])
				}
			}

			onAcquireR := func(msg sheetPrepareNotify) {}
			onModifyR := func(msg sheetModifyNotify) {
				logger.Debugf("[%+v] onModify", msg)
				msReceiver.Set(msg.Row, msg.Col, msg.Content)
				mapR.Store(cellKey{Row: msg.Row, Col: msg.Col}, msg.Content)
			}
			onReleaseR := func(msg sheetPrepareNotify) {}
			onConnR := func(msg sheetOnConnNotify) {
				logger.Debugf("[%+v] onConnection", msg)
				msReceiver = cache.NewMemSheetFromStringSlice(msg.Content, msg.Columns)
				for i := 0; i < len(msg.Content); i += 1 {
					mapR.Store(cellKey{Row: i/msg.Columns, Col: i%msg.Columns}, msg.Content[i])
				}
			}
			wsUrlS, wsUrlR := getWSAddr(tokenS, fid), getWSAddr(tokenR, fid)
			wsSender := NewWebSocket(t, wsUrlS, onAcquireS, onModifyS, onReleaseS, onConnS)


			runWS := func(ws *myWS, stopChan chan int) {
				for {
					select {
					case <- stopChan:
						return
					default:
						row, col := rand.Int()%100, rand.Int()%100
						content := strings.Repeat("test", rand.Int()%30)

						err := ws.SendJson("modify", sheetModifyMessage{
							Row:     row,
							Col:     col,
							Content: content,
							Info:    nil,
						})

						if !assert.NoError(t, err) {
							break
						}
						time.Sleep(pause)
					}
				}
			}
			stopChanS, stopChanR := make(chan int), make(chan int)
			go runWS(wsSender, stopChanS)
			time.Sleep(10 * time.Second)
			wsReceiver := NewWebSocket(t, wsUrlR, onAcquireR, onModifyR, onReleaseR, onConnR)
			go runWS(wsReceiver, stopChanR)
			time.Sleep(30 * time.Second)
			stopChanS <- 1
			stopChanR <- 1
			time.Sleep(15 * time.Second)
			ssS, ssR := msSender.ToStringSlice(), msReceiver.ToStringSlice()
			colS, rowS := msSender.Shape()
			colR, rowR := msReceiver.Shape()


			assert.Equal(t, colS, colR)
			assert.Equal(t, rowS, rowR)
			i := 0
			mapS.Range(func(k interface{}, v interface{}) bool {
				if u, ok := mapR.Load(k); assert.True(t, ok) {
					assert.Equalf(t, v, u, "S(%d, %d); R(%d, %d)", i/colS, i%colS, i/colR, i%colR)
				}
				i += 1
				return true
			})

			for i := 0; i < len(ssS); i += 1 {
				assert.Equalf(t, ssS[i], ssR[i], "S(%d, %d); R(%d, %d)", i/colS, i%colS, i/colR, i%colR)
			}

			wsSender.ws.Close()
			wsReceiver.ws.Close()
		}
	}
}

func goTestWorkFlowNoInterrupt(t *testing.T, pause time.Duration) {
	testWsNum := 20
	initRow, initCol := 100, 100
	tokenS, tokenR := login(t, loginParams[0]), login(t, loginParams[1])
	newSheetArg := utils.NewSheetParams{
		Token: tokenS,
		Name:  "test",
		InitRows: uint(initRow),
		InitColumns: uint(initCol),
	}
	raw, _ := getPostRaw(randomHostHttp(), "newsheet", newSheetArg)
	fidInt, err := strconv.Atoi(string(raw))
	fid := uint(fidInt)
	if assert.NoError(t, err) {
		getSheetArg := utils.GetSheetParams{
			Token: tokenR,
			Fid: fid,
		}
		getSheetRet := GetSheetRet{}
		err = getPostRet(randomHostHttp(), "getsheet", getSheetArg, &getSheetRet)
		if assert.NoError(t, err) {
			assert.EqualValues(t, fid, getSheetRet.Fid)
			assert.EqualValues(t, 0, getSheetRet.CheckPointNum)
			assert.Equal(t, "test", getSheetRet.Name)

			row, col := len(getSheetRet.Content) / getSheetRet.Columns, getSheetRet.Columns
			assert.Equal(t, len(getSheetRet.Content), row * col)

			ms := make([]*cache.MemSheet, testWsNum)
			for i := 0; i < testWsNum; i += 1 {
				ms[i] = cache.NewMemSheet(initRow, initCol)
			}

			var msSender, msReceiver *cache.MemSheet
			var mapS, mapR sync.Map

			type cellKey struct {
				Row int
				Col int
			}

			cnt := 0
			onAcquireS := func(msg sheetPrepareNotify) {}
			onModifyS := func(msg sheetModifyNotify) {
				logger.Info(cnt)
				cnt++
				logger.Debugf("[%+v] onModify", msg)
				msSender.Set(msg.Row, msg.Col, msg.Content)
				mapS.Store(cellKey{Row: msg.Row, Col: msg.Col}, msg.Content)
			}
			onReleaseS := func(msg sheetPrepareNotify) {}
			onConnS := func(msg sheetOnConnNotify) {
				logger.Debugf("[%+v] onConnection", msg)
				msSender = cache.NewMemSheetFromStringSlice(msg.Content, msg.Columns)
				for i := 0; i < len(msg.Content); i += 1 {
					mapS.Store(cellKey{Row: i/msg.Columns, Col: i%msg.Columns}, msg.Content[i])
				}
			}

			onAcquireR := func(msg sheetPrepareNotify) {}
			onModifyR := func(msg sheetModifyNotify) {
				logger.Debugf("[%+v] onModify", msg)
				msReceiver.Set(msg.Row, msg.Col, msg.Content)
				mapR.Store(cellKey{Row: msg.Row, Col: msg.Col}, msg.Content)
			}
			onReleaseR := func(msg sheetPrepareNotify) {}
			onConnR := func(msg sheetOnConnNotify) {
				logger.Debugf("[%+v] onConnection", msg)
				msReceiver = cache.NewMemSheetFromStringSlice(msg.Content, msg.Columns)
				for i := 0; i < len(msg.Content); i += 1 {
					mapR.Store(cellKey{Row: i/msg.Columns, Col: i%msg.Columns}, msg.Content[i])
				}
			}
			wsUrlS, wsUrlR := getWSAddr(tokenS, fid), getWSAddr(tokenR, fid)
			wsSender := NewWebSocket(t, wsUrlS, onAcquireS, onModifyS, onReleaseS, onConnS)
			wsReceiver := NewWebSocket(t, wsUrlR, onAcquireR, onModifyR, onReleaseR, onConnR)


			runWS := func(ws *myWS, stopChan chan int) {
				for {
					select {
					case <- stopChan:
						return
					default:
						row, col := rand.Int()%100, rand.Int()%100
						content := strings.Repeat("test", rand.Int()%30)

						err := ws.SendJson("modify", sheetModifyMessage{
							Row:     row,
							Col:     col,
							Content: content,
							Info:    nil,
						})

						if !assert.NoError(t, err) {
							break
						}
						time.Sleep(pause)
					}
				}
			}
			stopChanS, stopChanR := make(chan int), make(chan int)
			go runWS(wsSender, stopChanS)
			go runWS(wsReceiver, stopChanR)
			time.Sleep(30 * time.Second)
			stopChanS <- 1
			stopChanR <- 1
			time.Sleep(15 * time.Second)
			ssS, ssR := msSender.ToStringSlice(), msReceiver.ToStringSlice()
			colS, rowS := msSender.Shape()
			colR, rowR := msReceiver.Shape()


			assert.Equal(t, colS, colR)
			assert.Equal(t, rowS, rowR)
			i := 0
			mapS.Range(func(k interface{}, v interface{}) bool {
				if u, ok := mapR.Load(k); assert.True(t, ok) {
					assert.Equalf(t, v, u, "S(%d, %d); R(%d, %d)", i/colS, i%colS, i/colR, i%colR)
				}
				i += 1
				return true
			})

			for i := 0; i < len(ssS); i += 1 {
				assert.Equalf(t, ssS[i], ssR[i], "S(%d, %d); R(%d, %d)", i/colS, i%colS, i/colR, i%colR)
			}

			wsSender.ws.Close()
			wsReceiver.ws.Close()
		}
	}
}
