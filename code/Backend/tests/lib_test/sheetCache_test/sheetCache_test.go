package sheetCache_test

import (
	"backend/lib/cache"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMemSheet(t *testing.T) {
	// test regular shape MemSheet
	for p := 0; p < sizePower; p += 1 {
		memSheets[p] = getSizedMemSheet(1 << p, defaultRow, defaultCol)
	}
	for p := 0; p < sizePower; p += 1 {
		row, col := memSheets[p].Shape()
		assert.Equal(t, row, defaultRow)
		assert.Equal(t, col, defaultCol)
		assert.InEpsilon(t, int64((1 << p) << 20), memSheets[p].GetSize(), 0.01)
	}

	// test irregular shape MemSheet
	minCol, minRow := 100, 100
	for i := 0; i < 10; i += 1 {
		var ss []string
		var msNew, msNewSS *cache.MemSheet
		msNew = cache.NewMemSheet(minRow, minCol)
		row, col := rand.Int() % 100 + minRow, rand.Int() % 100 + minCol

		randCut := rand.Int() % col
		totRune := 0
		for j := 0; j < row * col - randCut; j += 1 {// test string slice that varies in (row*(col-1), row*col]
			runeNum := rand.Int() % (1 << 10)
			totRune += runeNum
			content := strings.Repeat("0", runeNum)
			ss = append(ss, content)
			msNew.Set(j / col, j % col, content)
		}
		msNewSS = cache.NewMemSheetFromStringSlice(ss, col)

		// size equal and in epsilon
		assert.Equal(t, msNew.GetSize(), msNewSS.GetSize())
		assert.InEpsilon(t, int64(totRune * 4), msNew.GetSize(), 0.01)

		// shape equal
		c1, r1 := msNew.Shape()
		c2, r2 := msNewSS.Shape()
		assert.Equal(t, c1, c2)
		assert.Equal(t, r1, r2)

		// content equal
		for j := 0; j < row * col; j += 1 {
			assert.Equal(t, msNew.Get(j / col, j % col), msNewSS.Get(j / col, j % col))
		}

		// test ToStringSlice equal
		msNewToSS := cache.NewMemSheetFromStringSlice(msNew.ToStringSlice(), col)
		for i := 0; i < row; i += 1 {
			for j := 0; j < col; j += 1 {
				assert.Equal(t, msNew.Get(i, j), msNewToSS.Get(i, j))
			}
		}
	}


	// test concurrent set, runeNum and get should be save
	wg := sync.WaitGroup{}
	workerN := 1000
	msCC := cache.NewMemSheet(minRow, minCol)
	testSet := []string{"test", "test_test", "test_test_test", "test_test_test_test", "test_test_test_test_test"}
	notifyChan := make(chan int)
	msCC.Set(0, 0, testSet[rand.Int() % len(testSet)])
	wg.Add(workerN)
	for i := 0; i < workerN; i += 1 {
		go func(idx int) {
			for {
				select {
				case <- notifyChan:
					msCC.Lock()
					content := msCC.Get(0, 0)
					assert.Equal(t, msCC.CellNet().RuneNum, int64(len(content)))
					assert.Contains(t, testSet, content)
					msCC.Unlock()
					t.Logf("[%d] worker finished", idx)
					wg.Done()
					return

				default:
					msCC.Set(0, 0, testSet[rand.Int() % len(testSet)])
				}
			}
		}(i)
	}
	for i := 0; i < workerN; i += 1 {
		time.Sleep(time.Millisecond)
		notifyChan <- 1
	}
	wg.Wait()

	// test concurrent set with expanding
}

func TestSheetCache(t *testing.T) {
	// test eviction
	MemoryCapByte := int64(MemoryCapMB << 20)
	sheetCache := cache.NewSheetCache(int64(MemoryCapByte))
	var keys []int
	for i := MemoryCapMB * 3; i >= 0; i -= 1 {
		ms1M := getSizedMemSheet(1, defaultRow, defaultCol)
		_, evictedKeys, _ := sheetCache.Add(i, ms1M)
		for _, key := range evictedKeys {
			keys = append(keys, key.(int))
		}
		k, e := sheetCache.Put(i)
		assert.Empty(t, k)
		assert.Empty(t, e)
		assert.IsDecreasing(t, keys)
	}

	sizeCnt := int64(0)
	unitSize := getSizedMemSheet(1, defaultRow, defaultCol).GetSize()
	for i := 0; i < MemoryCapMB * 3; i += 1 {
		sizeCnt += unitSize
		if sizeCnt <= MemoryCapByte {
			assert.NotNil(t, sheetCache.Get(i))
		} else {
			assert.Nil(t, sheetCache.Get(i))
		}
		sheetCache.Put(i)
	}

	ms := sheetCache.Get(0)
	row, col := ms.Shape()
	for i := 0; i < row * col; i += 1 {
		content := ms.Get(i / col, i % col)
		ms.Set(i / col, i % col, strings.Repeat(content, 4))
	}

	k, _ := sheetCache.Put(0)
	assert.ElementsMatch(t, []int{1, 2, 3}, k)

	// one sheet that cannot fit in cache
	msExcess := getSizedMemSheet(MemoryCapMB * 2, defaultRow, defaultCol)
	ms, _, _ = sheetCache.Add(-1, msExcess)
	assert.Nil(t, ms)

	for i := 1; i < MemoryCapMB * 3; i += 1 {
		sheetCache.Get(i)
	}	// 0 is not referred

	ms, _, _ = sheetCache.Add(-1, getSizedMemSheet(10, defaultRow, defaultCol))
	assert.Nil(t, ms)

	// test put existed MemSheet, but it becomes to big to fit in Cache
	ms = sheetCache.Get(0)
	for i := 0; i < row * col; i += 1 {
		content := ms.Get(i / col, i % col)
		ms.Set(i / col, i % col, strings.Repeat(content, MemoryCapMB))
	}
	k, e := sheetCache.Put(0)
	assert.ElementsMatch(t, []int{0}, k)
	assert.ElementsMatch(t, []*cache.MemSheet{ms}, e)
	// test 0 is deleted
	assert.Nil(t, sheetCache.Get(0))

	// test 0's memory is freed
	for i := 0; i <= 3; i += 1 {
		ms, _, _ = sheetCache.Add(i, getSizedMemSheet(1, defaultRow, defaultCol))
		assert.NotNil(t, ms)
	}

	// test 0's memory(4M) is freed correctly
	ms, _, _ = sheetCache.Add(4, getSizedMemSheet(1, defaultRow, defaultCol))
	assert.Nil(t, ms)
}