package cache

import (
	"backend/utils/logger"
	"sync"
	"unsafe"
)

// A 2d string slice with auto-scalability when Set is called on unbounded row and col
type CellNet struct {
	cells		[][]string
	maxRow		int
	maxCol		int

	RuneNum		int
}

func NewCellNet(initRow int, initCol int) *CellNet {
	cells := make([][]string, initRow)
	for i := 0; i < initRow; i += 1 {
		cells[i] = make([]string, initCol)
	}

	return &CellNet{
		cells: cells,
		maxRow: initRow,
		maxCol: initCol,

		RuneNum: 0,
	}
}

func NewCellNetFromStringSlice(ss []string, columns int) *CellNet {
	initRow := len(ss) / columns
	initCol := columns
	if initRow * columns < len(ss) {
		initRow += 1
		toExtend := make([]string, initRow * columns - len(ss))
		ss = append(ss, toExtend...)
	}
	cells := make([][]string, initRow)

	logger.Debug(initRow, initCol)

	runeNum := 0
	for i := 0; i < initRow; i += 1 {
		cells[i] = make([]string, initCol)
		for j := 0; j < initCol; j += 1 {
			cells[i][j] = ss[i * initCol + j]
			runeNum += len(cells[i][j])
		}
	}

	return &CellNet{
		cells: cells,
		maxRow: initRow,
		maxCol: initCol,

		RuneNum: runeNum,
	}
}

func (net *CellNet) Set(row int, col int, content string) {
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

func (net *CellNet) Get(row int, col int) string {
	if row > net.maxRow || col > net.maxCol {
		return ""
	}

	return net.cells[row][col]
}

func (net *CellNet) Shape() (rows int, cols int) {
	return net.maxRow, net.maxCol
}

func (net *CellNet) ToStringSlice() (ss []string) {
	cells := &net.cells

	for i := 0; i < len(*cells); i += 1 {
		ss = append(ss, (*cells)[i]...)
	}

	return ss
}


// In-memory sheet
type MemSheet struct {
	cells		*CellNet
}

func NewMemSheet(initRow int, initCol int) *MemSheet {
	return &MemSheet{
		cells: NewCellNet(initRow, initCol),
	}
}

func (ms *MemSheet) Set(row int, col int, content string) {
	if row < 0 || col < 0 {
		panic("row or column index is negative")
	}
	ms.cells.Set(row, col, content)
}

func (ms *MemSheet) Get(row int, col int) string {
	if row < 0 || col < 0 {
		panic("row or column index is negative")
	}
	return ms.cells.Get(row, col)
}

func (ms *MemSheet) GetSize() (size int) {
	size += int(unsafe.Sizeof(ms.cells.cells))
	size += ms.cells.maxRow * int(unsafe.Sizeof(ms.cells.cells[0]))
	size += ms.cells.maxRow * ms.cells.maxCol * int(unsafe.Sizeof(ms.cells.cells[0][0]))

	size += ms.cells.RuneNum * int(unsafe.Sizeof('0'))

	return
}

func (ms *MemSheet) Shape() (rows int, cols int) {
	return ms.cells.Shape()
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
func (sc *SheetCache) Add(key interface{}, ms *MemSheet) (evicted []*MemSheet) {
	evicted = nil
	if sc.curSize + int64(ms.GetSize()) > sc.maxSize {
		if evicted = sc.doEvict(int64(ms.GetSize())); evicted == nil {
			logger.Fatal("Cannot get enough memory from eviction!")
		}
	}

	sc.cache.Store(key, ms)
	return evicted
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

func (sc *SheetCache) doEvict(spareAtLeast int64) (evicted []*MemSheet) {
	if sc.curSize < spareAtLeast {
		return nil
	}

	// TODO: finish eviction, e.g. LRU

	return nil
}