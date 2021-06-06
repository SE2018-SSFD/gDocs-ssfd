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
	buf map[util.CacheID]Buffer
}

func InitCache() *Cache {
	c := &Cache{}
	return c
}

func (c *Cache) Get(cid util.CacheID) ([]byte, error) {
	c.Lock()
	buf := c.buf[cid].data
	c.Unlock()
	return buf, nil
}

func (c *Cache) Set(cid util.CacheID, buf []byte) error {
	c.Lock()
	c.buf[cid] = Buffer{data: buf}
	c.Unlock()
	return nil
}

func (c *Cache) Remove(cid util.CacheID) error {
	c.Lock()
	delete(c.buf, cid)
	c.Unlock()
	return nil
}

func (c *Cache) GetAndRemove(cid util.CacheID) ([]byte, error) {
	c.Lock()
	buf := c.buf[cid].data
	delete(c.buf, cid)
	c.Unlock()
	return buf, nil
}
