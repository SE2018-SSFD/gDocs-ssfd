package cache

import (
	"backend/lib/algorithm/lru"
	"backend/utils/logger"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"unsafe"
)

// A 2d string slice with auto-scalability when Set is called on unbounded row and col
type CellNet struct {
	cells		[][]*string
	maxRow		int32
	maxCol		int32
	gLock		sync.Mutex

	RuneNum		int64
}

func NewCellNet(initRow int32, initCol int32) *CellNet {
	cells := make([][]*string, initRow)
	for i := int32(0); i < initRow; i += 1 {
		cells[i] = make([]*string, initCol)
	}

	return &CellNet{
		cells: cells,
		maxRow: initRow,
		maxCol: initCol,

		RuneNum: 0,
	}
}

func NewCellNetFromStringSlice(ss []string, columns int32) *CellNet {
	initRow := int32(len(ss)) / columns
	initCol := columns
	if initRow * columns < int32(len(ss)) {
		initRow += 1
		toExtend := make([]string, initRow * columns - int32(len(ss)))
		ss = append(ss, toExtend...)
	}
	cells := make([][]*string, initRow)

	runeNum := int64(0)
	for i := int32(0); i < initRow; i += 1 {
		cells[i] = make([]*string, initCol)
		for j := int32(0); j < initCol; j += 1 {
			cells[i][j] = &ss[i * initCol + j]
			runeNum += int64(len(*cells[i][j]))
		}
	}

	return &CellNet{
		cells: cells,
		maxRow: initRow,
		maxCol: initCol,

		RuneNum: runeNum,
	}
}

func (net *CellNet) Set(row int32, col int32, content string) {
	net.gLock.Lock()
	if col + 1 > net.maxCol {
		for i := int32(0); i < net.maxRow; i += 1 {
			curRow := &net.cells[i]
			toExtendN := col + 1 - int32(len(*curRow))
			*curRow = append(*curRow, make([]*string, toExtendN)...)
		}

		net.maxCol = col + 1
	}

	if row + 1 > net.maxRow {
		toExtendN := row + 1 - net.maxRow
		toExtend := make([][]*string, toExtendN)
		for i := int32(0); i < toExtendN; i += 1 {
			toExtend[i] = make([]*string, net.maxCol)
		}

		net.cells = append(net.cells, toExtend...)
		net.maxRow = row + 1
	}
	net.gLock.Unlock()

	// add RuneNum and store new string atomically
	for {
		curCell := (*string)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&net.cells[row][col]))))
		if swapped := atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&net.cells[row][col])),
			unsafe.Pointer(curCell),
			unsafe.Pointer(&content)); swapped {
			if curCell == nil {
				atomic.AddInt64(&net.RuneNum, int64(len(content)))
			} else {
				atomic.AddInt64(&net.RuneNum, int64(len(content)) - int64(len(*curCell)))
			}
			break
		} else {
			continue
		}
	}
}

func (net *CellNet) Get(row int32, col int32) string {
	if row + 1 > net.maxRow || col + 1 > net.maxCol {
		return ""
	}

	if net.cells[row][col] == nil {
		return ""
	} else {
		return *net.cells[row][col]
	}
}

func (net *CellNet) Shape() (rows int32, cols int32) {
	return net.maxRow, net.maxCol
}

func (net *CellNet) ToStringSlice() (ss []string) {
	cells := &net.cells

	for i := int32(0); i < net.maxRow; i += 1 {
		toAppend :=  make([]string, net.maxCol)
		for j := int32(0); j < net.maxCol; j += 1 {
			k := (*cells)[i][j]
			if k == nil {
				toAppend[j] = ""
			} else {
				toAppend[j] = *k
			}
		}
		ss = append(ss, toAppend...)
	}

	return ss
}


// In-memory sheet
type MemSheet struct {
	cells		*CellNet
	monoLock	sync.RWMutex
	lastReport	int64
	refCount	int32
}

func NewMemSheet(initRow int, initCol int) *MemSheet {
	return &MemSheet{
		cells: NewCellNet(int32(initRow), int32(initCol)),
	}
}

func NewMemSheetFromStringSlice(ss []string, columns int) *MemSheet {
	return &MemSheet{
		cells: NewCellNetFromStringSlice(ss, int32(columns)),
	}
}

func (ms *MemSheet) Set(row int, col int, content string) {
	ms.monoLock.RLock()
	defer ms.monoLock.RUnlock()

	if row < 0 || col < 0 {
		logger.Warnf("[row:%d col:%d]row or column index is negative\n %s", row, col, debug.Stack())
		return
	}
	ms.cells.Set(int32(row), int32(col), content)
}

func (ms *MemSheet) Get(row int, col int) (content string) {
	if row < 0 || col < 0 {
		logger.Warnf("[row:%d col:%d]row or column index is negative\n %s", row, col, debug.Stack())
	}

	content = ms.cells.Get(int32(row), int32(col))

	return content
}

func (ms *MemSheet) refer() {
	atomic.AddInt32(&ms.refCount, 1)
}

func (ms *MemSheet) unRefer() {
	if refCnt := atomic.AddInt32(&ms.refCount, -1); refCnt < 0 {
		logger.Errorf("refCount of MemSheet is negative\n %s", debug.Stack())
	}
}

func (ms *MemSheet) isReferred() bool {
	return atomic.LoadInt32(&ms.refCount) > 0
}

func (ms *MemSheet) Lock() {
	ms.monoLock.Lock()
}

func (ms *MemSheet) Unlock() {
	ms.monoLock.Unlock()
}

func (ms *MemSheet) GetSize() int64 {
	size := int64(unsafe.Sizeof(ms.cells.cells))
	size += int64(ms.cells.maxRow * int32(unsafe.Sizeof(ms.cells.cells[0])))
	size += int64(ms.cells.maxRow * ms.cells.maxCol * int32(unsafe.Sizeof(ms.cells.cells[0][0])))

	size += ms.cells.RuneNum * int64(unsafe.Sizeof('0'))

	return size
}

func (ms *MemSheet) CellNet() *CellNet {
	return ms.cells
}

func (ms *MemSheet) reportSizeChange() (change int64) {
	change = ms.GetSize() - ms.lastReport
	ms.lastReport += change
	return change
}

func (ms *MemSheet) Shape() (rows int, cols int) {
	r, c := ms.cells.Shape()
	return int(r), int(c)
}

func (ms *MemSheet) ToStringSlice() (ss []string) {
	return ms.cells.ToStringSlice()
}


type SheetCache struct {
	maxSize			int64
	curSize			int64
	gLock			sync.RWMutex
	cache			sync.Map
	lru				*lru.LRU
}

func NewSheetCache(maxSize int64) *SheetCache {
	return &SheetCache{
		maxSize: maxSize,
		curSize: 0,
		lru: lru.NewLRU(),
	}
}

// If excess memory constraint, do eviction and return true
func (sc *SheetCache) Add(key interface{}, ms *MemSheet) (memSheet *MemSheet, keys []interface{}, evicted []*MemSheet) {
	spared, keys, evicted, doEvict, success := sc.doEvictIfNeeded(key, ms, false)
	logger.Infof("[%v] curSize: %d, maxSize: %d, toAdd: %d, doEvict: %t, evicted: %v, spared: %d",
		key, sc.curSize, sc.maxSize, ms.GetSize(), doEvict, keys, spared)

	if success {
		ms.refer()
		sc.lru.Add(key)
		sc.cache.Store(key, ms)
		return ms, keys, evicted
	} else {
		return nil, keys, evicted
	}
}

func (sc *SheetCache) Get(key interface{}) *MemSheet {
	sc.gLock.RLock()
	defer sc.gLock.RUnlock()

	if v, ok := sc.cache.Load(key); !ok {
		return nil
	} else {
		ms := v.(*MemSheet)
		ms.refer()
		sc.lru.Add(key)
		return ms
	}
}

func (sc *SheetCache) Put(key interface{}) (keys []interface{}, evicted []*MemSheet) {
	if v, ok := sc.cache.Load(key); !ok {
		return keys, evicted
	} else {
		ms := v.(*MemSheet)
		_, keys, evicted, doEvict, success := sc.doEvictIfNeeded(key, ms, true)
		if !doEvict && !success {
			return keys, evicted
		} else {
			ms.unRefer()
			return keys, evicted
		}
	}
}

func (sc *SheetCache) doEvictIfNeeded(key interface{}, ms *MemSheet, isPut bool) (spared int64,
	keys []interface{}, evicted []*MemSheet, doEvict bool, success bool) {
	sc.gLock.Lock()
	defer sc.gLock.Unlock()

	ms.Lock()
	defer ms.Unlock()

	changedSize := int64(0)
	if isPut {
		if _, ok := sc.cache.Load(key); ok {
			changedSize = ms.reportSizeChange()
		} else {// in case that, put the same sheet more than once, which existed in cache but cannot fit anymore
			return 0, []interface{}{}, []*MemSheet{}, false, false
		}
	} else {
		if v, ok := sc.cache.Load(key); ok {
			old := v.(*MemSheet)
			changedSize = ms.GetSize() - old.GetSize()
		} else {
			changedSize = ms.GetSize()
		}
	}

	if sc.curSize + changedSize > sc.maxSize {
		if keys, evicted, spared = sc.doEvict(key, changedSize - (sc.maxSize - sc.curSize)); spared == 0 {
			if sc.maxSize != 0 {
				logger.Warnf("[%v] Cannot get enough memory from eviction! %s", key, debug.Stack())
			}
			if isPut {
				sc.lru.Delete(key)
				sc.cache.Delete(key)
				atomic.AddInt64(&sc.curSize, -(ms.GetSize() - changedSize))
			}
			return 0, []interface{}{key}, []*MemSheet{ms}, true, false
		} else {
			sc.curSize += changedSize - spared
			ms.reportSizeChange()
			return spared, keys, evicted, true, true
		}
	} else {
		sc.curSize += changedSize
		ms.reportSizeChange()
		return 0, []interface{}{}, []*MemSheet{}, false, true
	}
}

func (sc *SheetCache) doEvict(caller interface{}, spareAtLeast int64) (keys []interface{}, evicted []*MemSheet, spared int64) {
	if sc.curSize < spareAtLeast {
		return nil, nil, 0
	}

	noEvictionCnt := int64(0)
	for spared < spareAtLeast {
		if sc.lru.Len() == 0 {
			redoEviction(sc, keys, evicted)
			return nil, nil, 0
		}
		toEvictKey := sc.lru.DoEvict()
		if toEvictKey == caller {
			sc.lru.Add(toEvictKey)
			noEvictionCnt += spareAtLeast
		} else {
			if v, ok := sc.cache.Load(toEvictKey); !ok {
				logger.Fatalf("cache: memSheet in lru but not in cache!\n %s", debug.Stack())
			} else {
				toEvict := v.(*MemSheet)
				toEvict.Lock()
				if toEvict.isReferred() {
					sc.lru.Add(toEvictKey)
					noEvictionCnt += toEvict.GetSize()
					logger.Debugf("[%v] is referred", toEvictKey)
				} else {
					sc.cache.Delete(toEvictKey)
					spared += toEvict.GetSize()
					keys = append(keys, toEvictKey)
					evicted = append(evicted, toEvict)
					noEvictionCnt = 0
					logger.Debugf("[%v] can be evicted, spared %d", toEvictKey, spared)
				}
				toEvict.Unlock()
			}
		}

		// searching for evictable MemSheet (not referred) for 3 cycles and evict nothing
		if noEvictionCnt > 3 * sc.curSize {
			redoEviction(sc, keys, evicted)
			return nil, nil, 0
		}
	}

	return keys, evicted, spared
}

func redoEviction(sc *SheetCache, keys []interface{}, evicted []*MemSheet) {
	for i := len(keys) - 1; i >= 0; i -= 1 {
		evicted[i].Lock()
		sc.cache.Store(keys[i], evicted[i])
		sc.lru.AddToLeastRecent(keys[i])
		evicted[i].Unlock()
	}
}