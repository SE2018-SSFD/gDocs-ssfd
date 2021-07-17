package sheetCache_test

import (
	"backend/lib/cache"
	loggerWrap "backend/utils/logger"
	"github.com/kataras/golog"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"
)

const (
	sizePower	=	6
	defaultRow	=	1 << 5
	defaultCol	=	1 << 5
	MemoryCapMB	=	64
)

var (
	memSheets [sizePower]*cache.MemSheet
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	loggerWrap.SetLogger(golog.New())
	loggerWrap.SetLevel("Debug")

	exitCode := m.Run()
	os.Exit(exitCode)
}

func getSizedMemSheet(mb int, row int, col int) *cache.MemSheet {
	ss := make([]string, row * col)
	for i := 0; i < row * col; i += 1 {
		ss[i] = strings.Repeat("0", mb * 1024 / int(unsafe.Sizeof('0')))
	}
	return cache.NewMemSheetFromStringSlice(ss, col)
}