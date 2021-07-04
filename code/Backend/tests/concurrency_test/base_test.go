package concurrency_test

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

var (
	hosts = []string{
		"localhost:10086",
		"localhost:10087",
		"localhost:10088",
	}
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	exitCode := m.Run()
	os.Exit(exitCode)
}

func randomHost() string {
	idx := rand.Int() % 3
	return hosts[idx]
}