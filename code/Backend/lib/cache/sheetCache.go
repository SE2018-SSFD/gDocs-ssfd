package cache

import (
	"sync"
	"unsafe"
)

// A 2d string slice with auto-scalability when set is called on unbounded row and col
type cellNet struct {
	cells		[][]string
	maxRow		int
	maxCol		int

	RuneNum		int
}

func newCellNet(initRow int, initCol int) *cellNet {
	cells := make([][]string, initRow)
	for i := 0; i < initRow; i += 1 {
		cells[i] = make([]string, initCol)
	}

	return &cellNet{
		cells: cells,
		maxRow: initRow,
		maxCol: initCol,

		RuneNum: 0,
	}
}

func (net *cellNet) set(row int, col int, content string) {
	if col + 1 > net.maxCol {
		for i := 0; i < net.maxRow; i += 1 {
			curRow := &net.cells[i]
			toExtendN := col + 1 - len(*curRow)
			*curRow = append(*curRow, make([]string, toExtendN)...)
		}

		net.maxCol = col + 1
	}

	if row + 1 > net.maxRow {
		toExtendN := row + 1 - net.maxRow
		toExtend := make([][]string, toExtendN)
		for i := 0; i < toExtendN; i += 1 {
			toExtend[i] = make([]string, net.maxCol)
		}

		net.cells = append(net.cells, toExtend...)
		net.maxRow = row + 1
	}

	net.RuneNum += len(content) - len(net.cells[row][col])
	net.cells[row][col] = content
}

func (net *cellNet) get(row int, col int) string {
	if row > net.maxRow || col > net.maxCol {
		return ""
	}

	return net.cells[row][col]
}


// In-memory sheet
type MemSheet struct {
	cells		*cellNet
}

func NewMemSheet(initRow int, initCol int) *MemSheet {
	return &MemSheet{
		cells: newCellNet(initRow, initCol),
	}
}

func (ms *MemSheet) Set(row int, col int, content string) {
	if row < 0 || col < 0 {
		panic("row or column index is negative")
	}
	ms.cells.set(row, col, content)
}

func (ms *MemSheet) Get(row int, col int) string {
	if row < 0 || col < 0 {
		panic("row or column index is negative")
	}
	return ms.cells.get(row, col)
}

func (ms *MemSheet) GetSize() (size int) {
	size += int(unsafe.Sizeof(ms.cells.cells))
	size += ms.cells.maxRow * int(unsafe.Sizeof(ms.cells.cells[0]))
	size += ms.cells.maxRow * ms.cells.maxCol * int(unsafe.Sizeof(ms.cells.cells[0][0]))

	size += ms.cells.RuneNum * int(unsafe.Sizeof('0'))

	return
}

func (ms *MemSheet) Shape() (int, int) {
	return ms.cells.maxRow, ms.cells.maxCol
}

func (ms *MemSheet) ToStringSlice() (ss []string) {
	cells := &ms.cells.cells

	for i := 0; i < len(*cells); i += 1 {
		ss = append(ss, (*cells)[i]...)
	}

	return
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
func (sc *SheetCache) Add(key interface{}, ms *MemSheet) (evicted bool) {
	evicted = false
	if sc.curSize + int64(ms.GetSize()) > sc.maxSize {
		// TODO: evict
		evicted = true
	}

	sc.cache.Store(key, ms)
	return
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

func (sc *SheetCache) doEvict(spareAtLeast int64) bool {
	if sc.curSize < spareAtLeast {
		return false
	}

	// TODO: finish eviction, e.g. LRU

	return true
}