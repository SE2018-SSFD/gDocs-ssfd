package cache

import (
	"backend/utils/logger"
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

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&net.cells[row][col])), unsafe.Pointer(&content))
	atomic.AddInt64(&net.RuneNum, int64(len(content)) - int64(len(*net.cells[row][col])))
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
	monoLock	sync.Mutex
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
	if row < 0 || col < 0 {
		panic("row or column index is negative")
	}
	ms.cells.Set(int32(row), int32(col), content)
}

func (ms *MemSheet) Get(row int, col int) (content string) {
	if row < 0 || col < 0 {
		panic("row or column index is negative")
	}

	content = ms.cells.Get(int32(row), int32(col))

	return content
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
	cache			sync.Map
}

func NewSheetCache(maxSize int64) *SheetCache {
	return &SheetCache{
		maxSize: maxSize,
		curSize: 0,
	}
}

// If excess memory constraint, do eviction and return true
func (sc *SheetCache) Add(key interface{}, ms *MemSheet) (keys []interface{}, evicted []*MemSheet) {
	if sc.curSize + ms.GetSize() > sc.maxSize {
		if keys, evicted = sc.doEvict(ms.GetSize()); evicted == nil {
			logger.Error("Cannot get enough memory from eviction!")
			return nil, nil
		}
	}

	sc.cache.Store(key, ms)
	return keys, evicted
}

func (sc *SheetCache) Get(key interface{}) *MemSheet {
	if v, ok := sc.cache.Load(key); !ok {
		return nil
	} else {
		return v.(*MemSheet)
	}
}

func (sc *SheetCache) Del(key interface{}) {
	sc.cache.Delete(key)
}

func (sc *SheetCache) doEvict(spareAtLeast int64) (keys []interface{}, evicted []*MemSheet) {
	if sc.curSize < spareAtLeast {
		return nil, nil
	}

	// TODO: finish eviction, e.g. LRU

	return nil, nil
}