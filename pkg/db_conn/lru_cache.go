package db_conn

import (
	"container/list"
	"sync"
)

type cacheEntry[T any] struct {
	key  string
	data *T
	elem *list.Element
}

type LruCache[T any] struct {
	mu         sync.RWMutex
	ll         *list.List
	cache      map[string]cacheEntry[T]
	maxEntries int
}

func NewLruCache[T any](maxEntries int) *LruCache[T] {
	return &LruCache[T]{
		ll:         list.New(),
		cache:      make(map[string]cacheEntry[T]),
		maxEntries: maxEntries,
	}
}

func (c *LruCache[T]) Get(key string) (*T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ent, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ent.elem)
		return ent.data, true
	}
	return nil, false
}

func (c *LruCache[T]) Set(key string, data *T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ent.elem)
		ent.data = data
		return
	}

	ent := cacheEntry[T]{key: key, data: data, elem: c.ll.PushFront(data)}
	ent.elem = c.ll.PushFront(data)
	c.cache[key] = ent

	if c.maxEntries != 0 && c.ll.Len() > c.maxEntries {
		c.removeOldest()
	}
}

func (c *LruCache[T]) Delete(key string) {
	if ent, ok := c.cache[key]; ok {
		c.removeElem(&ent)
	}
}

func (c *LruCache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll = list.New()
	c.cache = make(map[string]cacheEntry[T])
}

func (c *LruCache[T]) removeElem(ent *cacheEntry[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll.Remove(ent.elem)
	delete(c.cache, ent.key)
}

func (c *LruCache[T]) removeOldest() {
	if ele := c.ll.Back(); ele != nil {
		c.removeElem(ele.Value.(*cacheEntry[T]))
	}
}

func (c *LruCache[T]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ll.Len()
}
