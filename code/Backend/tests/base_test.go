package tests

import (
	"backend/lib/zkWrap"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
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