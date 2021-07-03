package zkWrap_test

import (
	"backend/lib/zkWrap"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHeartbeat(t *testing.T) {
	var hbs []*zkWrap.Heartbeat
	var hbMeCntConnLock, hbWhoCntConnLock, hbMeCntDisConnLock, hbWhoCntDisConnLock sync.RWMutex
	hbMeCntConn, hbWhoCntConn := make(map[string]int), make(map[string]int)
	hbMeCntDisConn, hbWhoCntDisConn := make(map[string]int), make(map[string]int)
	hosts := []string{"127.0.0.1:1111", "127.0.0.1:2222", "127.0.0.1:3333", "127.0.0.1:4444"}

	onConn := func (me string, who string) {
		hbMeCntConnLock.Lock(); hbMeCntConn[me] += 1; hbMeCntConnLock.Unlock()
		hbWhoCntConnLock.Lock(); hbWhoCntConn[who] += 1; hbWhoCntConnLock.Unlock()
		t.Log(me, "onConn", who)
	}
	onDisConn := func (me string, who string) {
		hbMeCntDisConnLock.Lock(); hbMeCntDisConn[me] += 1; hbMeCntDisConnLock.Unlock()
		hbWhoCntDisConnLock.Lock(); hbWhoCntDisConn[who] += 1; hbWhoCntDisConnLock.Unlock()
		println(me, "onDisConn", who)
	}

	for idx, addr := range hosts {
		hb, err := zkWrap.RegisterHeartbeat("test", time.Second * 15, addr, onConn, onDisConn)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}
		time.Sleep(time.Second)

		assert.ElementsMatch(t, hb.GetOriginMates(), hosts[:idx])

		hbs = append(hbs, hb)
	}

	time.Sleep(time.Second)

	for idx, addr := range hosts {
		hbMeCntConnLock.RLock(); assert.Equal(t, hbMeCntConn[addr], len(hosts) - idx); hbMeCntConnLock.RUnlock()
		hbWhoCntConnLock.RLock(); assert.Equal(t, hbWhoCntConn[addr], idx + 1); hbWhoCntConnLock.RUnlock()
	}

	for _, hb := range hbs {
		hb.Disconnect()
		time.Sleep(time.Second)
	}

	for idx, addr := range hosts {
		hbMeCntDisConnLock.RLock(); assert.Equal(t, hbMeCntDisConn[addr], idx); hbMeCntDisConnLock.RUnlock()
		hbWhoCntDisConnLock.RLock(); assert.Equal(t, hbWhoCntDisConn[addr], len(hosts) - idx - 1); hbWhoCntDisConnLock.RUnlock()
	}
}

func TestElection(t *testing.T) {
	testN := 10
	notifyChan := make(chan int, testN)
	for i := 0; i < testN; i += 1 {
		cb := func(el *zkWrap.Elector) {
			println(el.Me, "is leader !!!!!!!")
			assert.Equal(t, el.IsLeader, true)
			assert.Equal(t, el.IsRunning, true)

			el.Resign()
			assert.Equal(t, el.IsLeader, false)
			assert.Equal(t, el.IsRunning, true)

			el.StopElection()
			assert.Equal(t, el.IsLeader, false)
			assert.Equal(t, el.IsRunning, false)

			notifyChan <- 1
		}

		_, err := zkWrap.NewElector("test", string(rune('A' + i)), cb)
		if err != nil {
			t.Fail()
		}
	}

	stopTimerChan := make(chan int)
	go func() {
		select {
		case <- time.After(time.Second * 100):
			t.Error("election test timout")
			return
		case <-stopTimerChan:
			return
		}
	}()

	waitOnChanN(notifyChan, testN)
	stopTimerChan <- 1
}

func TestMutex(t *testing.T) {
	testN := 10
	notifyChan := make(chan int, testN)

	refMutex := sync.Mutex{}
	lockedN := 0

	testWorker := func(c chan int, lockName string, workerID string, lockTime time.Duration) {
		l, _ := zkWrap.NewMutex(lockName)

		t.Log(workerID, "\tLocking...\t", time.Now())
		if err := l.Lock(); err != nil {
			t.Fail()
		}

		refMutex.Lock()
		assert.Equal(t, lockedN, 0)
		lockedN += 1
		refMutex.Unlock()
		t.Log(workerID, "\tLocked!\t\t", "locked by", lockedN, "\t", time.Now())

		time.Sleep(lockTime)

		refMutex.Lock()
		assert.Equal(t, lockedN, 1)
		lockedN -= 1
		refMutex.Unlock()
		t.Log(workerID, "\tUnlocked!\t", "locked by", lockedN, "\t", time.Now())
		if err := l.Unlock(); err != nil {
			t.Fail()
		}

		c <- 1
	}

	for i := 0; i < testN; i += 1 {
		go testWorker(notifyChan, "testlock", string(rune('A' + i)), time.Second)
	}

	waitOnChanN(notifyChan, testN)
}

func TestLogRoom(t *testing.T) {
	var testLogRoomName = "test7777777"
	var testRoomN = 10
	var testAppendLogNum = 5
	var testOriginLogNum = 20
	var zkLogRooms = make([]*zkWrap.LogRoom, testRoomN)

	var testOriginChanNum = 20
	var testNewChanNum = 5

	var wgTotalRoom = sync.WaitGroup{}
	var wgsRoom = make([]sync.WaitGroup, testRoomN)

	var wgTotalLog = sync.WaitGroup{}

	var err error

	wgTotalRoom.Add(testNewChanNum * testRoomN + testOriginChanNum)
	var cntRoom int32 = 0
	onNewChannelCallback := func(cid int) {
		atomic.AddInt32(&cntRoom, 1)
		t.Logf("[%d %d] onNewChannelCallback", cid, cntRoom)
		wgTotalRoom.Done()
	}

	wgTotalLog.Add(testAppendLogNum * testRoomN + testOriginLogNum)
	var cntLog int32 = 0
	onAppendCallback := func(log zkWrap.LogNode) {
		atomic.AddInt32(&cntLog, 1)
		t.Logf("[%d %s %d] onAppendCallback", log.Lid, log.Content, cntLog)
		wgTotalLog.Done()
	}

	var originChanNum int
	if zkLogRooms[0], originChanNum, err = zkWrap.NewLogRoom(testLogRoomName, onNewChannelCallback, onAppendCallback); err != nil {
		t.Fatalf("%+v", errors.WithStack(err))
	} else {
		assert.Equal(t, 0, originChanNum)
	}

	for i := 0; i < testOriginChanNum; i += 1 {
		if cid, err := zkLogRooms[0].NewLogChannel(); err != nil {
			t.Fatalf("%d %+v", i, errors.WithStack(err))
		} else {
			assert.Equal(t, i, cid)
		}
	}

	for i := 0; i < testOriginLogNum; i += 1 {
		 if lid, err := zkLogRooms[0].GetLogChannel(0).Append("testOriginLog_" + strconv.Itoa(i)); err != nil {
			 t.Fatalf("%+v", errors.WithStack(err))
		 } else {
		 	assert.Equal(t, i, lid)
		 }
	}

	for i := 1; i < testRoomN; i += 1 {
		wgsRoom[i].Add(testOriginChanNum)
	}

	stopTimerChan := make(chan int, 0)
	go func() {
		select {
		case <- time.After(time.Second * 100):
			t.Fatal("originLogChan test timout")
			return
		case <-stopTimerChan:
			return
		}
	}()

	for i := 1; i < testRoomN; i += 1 {
		if zkLogRooms[i], originChanNum, err = zkWrap.NewLogRoom(testLogRoomName, onNewChannelCallback, onAppendCallback); err != nil {
			t.Fatalf("%+v", errors.WithStack(err))
		}
		t.Log("originChanNum", originChanNum)
		for j := 0;  j < originChanNum; j += 1 {
			wgsRoom[i].Done()
		}
	}

	for i := 1; i < testRoomN; i += 1 {
		wgsRoom[i].Wait()
		t.Log("waiting for origin Chan Success", i)
	}
	stopTimerChan <- 1

	for i := 0; i < testNewChanNum; i += 1 {
		t.Log("testNewChan i:", i)
		if cid, err := zkLogRooms[0].NewLogChannel(); err != nil {
			t.Fatalf("%+v", errors.WithStack(err))
		} else {
			assert.Equal(t, testOriginChanNum+i, cid)
		}
	}

	go func() {
		select {
		case <- time.After(time.Second * 100):
			t.Fatal("newLogChan test timout")
			return
		case <-stopTimerChan:
			return
		}
	}()

	wgTotalRoom.Wait()
	stopTimerChan <- 1

	// test originLog
	for i := 1; i < testRoomN; i += 1 {
		assert.Equal(t, testOriginLogNum, len(zkLogRooms[i].GetLogChannel(0).GetOriginLog()))
	}

	// test appendLog
	for i := 0; i < testAppendLogNum; i += 1 {
		if lid, err := zkLogRooms[0].GetLogChannel(0).Append("testAppendLog_" + strconv.Itoa(i)); err != nil {
			t.Fatalf("%+v", errors.WithStack(err))
		} else {
			assert.Equal(t, testOriginLogNum+i, lid)
		}
	}

	go func() {
		select {
		case <- time.After(time.Second * 100):
			t.Fatal("newLogChan test timout")
			return
		case <-stopTimerChan:
			return
		}
	}()

	wgTotalLog.Wait()
	stopTimerChan <- 1

	_ = zkLogRooms[0].GetLogChannel(0).DeleteAll()
	checkLogEmpty, _ := zkWrap.NewLog(testLogRoomName + "/" + "logRoom0000000000", onAppendCallback)
	assert.Equal(t, 0, len(checkLogEmpty.GetOriginLog()))

	_ = zkWrap.ClearLogRoom(testLogRoomName)
	//for i := 0; i < testRoomN; i += 1 {
	//	zkLogRooms[i].DisConnect()
	//}
	//_, checkRoomEmpty, _ := zkWrap.NewLogRoom(testLogRoomName, onNewChannelCallback, onAppendCallback)
	//assert.Equal(t, 0, checkRoomEmpty)
}