package dfs_test

import (
	"backend/utils/logger"
	"github.com/kataras/golog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger.SetLogger(golog.New())
	logger.SetLevel("Debug")
	exitCode := m.Run()
	os.Exit(exitCode)
}