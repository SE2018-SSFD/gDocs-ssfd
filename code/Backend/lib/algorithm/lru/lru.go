package lru

import (
	"container/list"
	"sync"
)

type LRU struct {
	lock			sync.Mutex
	hashMap			map[interface{}]*list.Element
	cacheList		*list.List
}

func NewLRU() *LRU {
	return &LRU{
		hashMap: make(map[interface{}]*list.Element),
		cacheList: list.New(),
	}
}

func (lru *LRU) Add(item interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if elem := lru.hashMap[item]; elem == nil {
		elem = lru.cacheList.PushFront(item)
		lru.hashMap[item] = elem
	} else {
		lru.cacheList.MoveToFront(elem)
	}
}

func (lru *LRU) AddToLeastRecent(item interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if elem := lru.hashMap[item]; elem == nil {
		elem = lru.cacheList.PushBack(item)
		lru.hashMap[item] = elem
	} else {
		lru.cacheList.MoveToBack(elem)
	}
}

func (lru *LRU) Delete(item interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if elem := lru.hashMap[item]; elem == nil {
		return
	} else {
		delete(lru.hashMap, item)
		lru.cacheList.Remove(elem)
	}
}

func (lru *LRU) DoEvict() (item interface{}) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	evict := lru.cacheList.Back()
	lru.cacheList.Remove(evict)
	item = evict.Value
	delete(lru.hashMap, item)

	return item
}

func (lru *LRU) Len() int {
	return lru.cacheList.Len()
}