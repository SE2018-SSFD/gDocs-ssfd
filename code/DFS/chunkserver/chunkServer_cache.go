package chunkserver

import (
	"DFS/util"
	"sync"
)

type Buffer struct {
	data []byte
}

type Cache struct {
	sync.RWMutex
	buf map[util.Handle]Buffer
}

func InitCache() *Cache {
	c := &Cache{}
	return c
}

func (c *Cache) Get(handle util.Handle, buf []byte) error {
	c.Lock()
	buf = c.buf[handle].data
	c.Unlock()
	return nil
}

func (c *Cache) Set(handle util.Handle, buf []byte) error {
	c.Lock()
	c.buf[handle] = Buffer{data: buf}
	c.Unlock()
	return nil
}
