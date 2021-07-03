package lru_test

import (
	"backend/lib/algorithm/lru"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	input = []uint{1, 2, 3, 4, 5, 1, 2, 6, 7, 8, 9, 4, 1}
	expect = []uint{3, 5, 2, 6, 7, 8, 9, 4, 1}
)

func TestLRU(t *testing.T) {
	l := lru.NewLRU()
	for _, in := range input {
		l.Add(in)
	}

	for _, exp := range expect {
		assert.Equal(t, exp, l.DoEvict().(uint))
	}
}
