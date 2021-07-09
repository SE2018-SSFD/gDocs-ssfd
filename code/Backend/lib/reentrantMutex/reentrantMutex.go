package reentrantMutex

import (
	"backend/utils/logger"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

type ReentrantMutex struct {
	mu        *sync.Mutex
	cond      *sync.Cond
	owner     int
	HoldCount int32
}

func NewReentrantMutex() *ReentrantMutex {
	rl := &ReentrantMutex{}
	rl.mu = new(sync.Mutex)
	rl.cond = sync.NewCond(rl.mu)
	return rl
}

func GetGoroutineId() int {
	defer func()  {
		if err := recover(); err != nil {
			logger.Errorf("Recovered: Fail in GetGoroutineId()\n%s", debug.Stack())
		}
	}()

	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func (rl *ReentrantMutex) Lock() {
	me := GetGoroutineId()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.owner == me {
		rl.HoldCount++
		return
	}
	for rl.HoldCount != 0 {
		rl.cond.Wait()
	}
	rl.owner = me
	rl.HoldCount = 1
}

func (rl *ReentrantMutex) Unlock() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.HoldCount == 0 || rl.owner != GetGoroutineId() {
		logger.Errorf("[goId(%d)] Trying to unlock a reentrantMutex that doesn't belong to this goroutine",
			GetGoroutineId())
	}
	rl.HoldCount--
	if rl.HoldCount == 0 {
		rl.cond.Signal()
	}
}
