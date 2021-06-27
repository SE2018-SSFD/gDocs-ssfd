package test
//import (
//	"DFS/util/zkWrap"
//	"github.com/stretchr/testify/assert"
//	"sync"
//	"testing"
//	"time"
//)
//
//func TestHeartbeat(t *testing.T) {
//	var hbs []*zkWrap.Heartbeat
//	hbMeCntConn, hbWhoCntConn := make(map[string]int), make(map[string]int)
//	hbMeCntDisConn, hbWhoCntDisConn := make(map[string]int), make(map[string]int)
//	hosts := []string{"127.0.0.1:1111", "127.0.0.1:2222", "127.0.0.1:3333", "127.0.0.1:4444"}
//
//	onConn := func (me string, who string) {
//		hbMeCntConn[me] += 1
//		hbWhoCntConn[who] += 1
//		t.Log(me, "onConn", who)
//	}
//	onDisConn := func (me string, who string) {
//		hbMeCntDisConn[me] += 1
//		hbWhoCntDisConn[who] += 1
//		println(me, "onDisConn", who)
//	}
//
//	for idx, addr := range hosts {
//		hb, err := zkWrap.RegisterHeartbeat("test", time.Second * 15, addr, onConn, onDisConn)
//		if err != nil {
//			t.Fail()
//		}
//		time.Sleep(time.Second)
//
//		assert.ElementsMatch(t, hb.GetMates(), hosts[:idx])
//
//		hbs = append(hbs, hb)
//	}
//
//	time.Sleep(time.Second)
//
//	for idx, addr := range hosts {
//		assert.Equal(t, hbMeCntConn[addr], len(hosts) - idx)
//		assert.Equal(t, hbWhoCntConn[addr], idx + 1)
//	}
//
//	for _, hb := range hbs {
//		hb.Disconnect()
//		time.Sleep(time.Second)
//	}
//
//	for idx, addr := range hosts {
//		assert.Equal(t, hbMeCntDisConn[addr], idx)
//		assert.Equal(t, hbWhoCntDisConn[addr], len(hosts) - idx - 1)
//	}
//}
//
//func TestElection(t *testing.T) {
//	testN := 10
//	notifyChan := make(chan int, testN)
//	for i := 0; i < testN; i += 1 {
//		cb := func(el *zkWrap.Elector) {
//			println(el.Me, "is leader !!!!!!!")
//			assert.Equal(t, el.IsLeader, true)
//			assert.Equal(t, el.IsRunning, true)
//
//			el.Resign()
//			assert.Equal(t, el.IsLeader, false)
//			assert.Equal(t, el.IsRunning, true)
//
//			el.StopElection()
//			assert.Equal(t, el.IsLeader, false)
//			assert.Equal(t, el.IsRunning, false)
//
//			notifyChan <- 1
//		}
//
//		_, err := zkWrap.NewElector("test", string(rune('A' + i)), cb)
//		if err != nil {
//			t.Fail()
//		}
//	}
//
//	stopTimerChan := make(chan int)
//	go func() {
//		select {
//		case <- time.After(time.Second * 100):
//			t.Error("election test timout")
//			return
//		case <-stopTimerChan:
//			return
//		}
//	}()
//
//	waitOnChanN(notifyChan, testN)
//	stopTimerChan <- 1
//}
//
//func TestMutex(t *testing.T) {
//	testN := 10
//	notifyChan := make(chan int, testN)
//
//	refMutex := sync.Mutex{}
//	lockedN := 0
//
//	testWorker := func(c chan int, lockName string, workerID string, lockTime time.Duration) {
//		l, _ := zkWrap.NewMutex(lockName)
//
//		t.Log(workerID, "\tLocking...\t", time.Now())
//		if err := l.Lock(); err != nil {
//			t.Fail()
//		}
//
//		refMutex.Lock()
//		assert.Equal(t, lockedN, 0)
//		lockedN += 1
//		refMutex.Unlock()
//		t.Log(workerID, "\tLocked!\t\t", "locked by", lockedN, "\t", time.Now())
//
//		time.Sleep(lockTime)
//
//		refMutex.Lock()
//		assert.Equal(t, lockedN, 1)
//		lockedN -= 1
//		refMutex.Unlock()
//		t.Log(workerID, "\tUnlocked!\t", "locked by", lockedN, "\t", time.Now())
//		if err := l.Unlock(); err != nil {
//			t.Fail()
//		}
//
//		c <- 1
//	}
//
//	for i := 0; i < testN; i += 1 {
//		go testWorker(notifyChan, "testlock", string(rune('A' + i)), time.Second)
//	}
//
//	waitOnChanN(notifyChan, testN)
//}
