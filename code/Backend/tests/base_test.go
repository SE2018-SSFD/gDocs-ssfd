package tests

import (
	"backend/lib/zkWrap"
	"backend/utils/config"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config.Get().ZKAddr = "localhost:12086;localhost:12087;localhost:12088"
	err := zkWrap.Chroot("/test")
	if err != nil {
		os.Exit(-1)
	}
	exitCode := m.Run()
	os.Exit(exitCode)
}

func waitOnChanN(c chan int, n int) {
	for i := 0; i < n; i += 1 {
		_ = <- c
	}
}