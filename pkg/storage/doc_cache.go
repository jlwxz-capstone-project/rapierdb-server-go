package storage

import (
	"container/list"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

const DEFAULT_CACHE_SIZE = 1024 // 默认缓存大小

type LoadedDoc struct {
	DocID string
	Doc   *loro.LoroDoc
}

type cacheEntry struct {
	key  string
	doc  *LoadedDoc
	elem *list.Element
}

type DocsCache struct {
	maxEntries int
	ll         *list.List             // LRU 链表
	cache      map[string]*cacheEntry // 缓存映射
}

func NewEmptyDocsCache() *DocsCache {
	return &DocsCache{
		maxEntries: DEFAULT_CACHE_SIZE,
		ll:         list.New(),
		cache:      make(map[string]*cacheEntry),
	}
}

func (c *DocsCache) Get(key []byte) (*loro.LoroDoc, bool) {
	if ent, ok := c.cache[util.Bytes2String(key)]; ok {
		c.ll.MoveToFront(ent.elem) // 移到链表头部
		return ent.doc.Doc, true
	}
	return nil, false
}

func (c *DocsCache) Set(key []byte, docID string, doc *loro.LoroDoc) {
	keyStr := util.Bytes2String(key)
	if ent, ok := c.cache[keyStr]; ok {
		// 已存在则更新
		c.ll.MoveToFront(ent.elem)
		ent.doc = &LoadedDoc{DocID: docID, Doc: doc}
		return
	}

	// 新增条目
	ent := &cacheEntry{
		key: keyStr,
		doc: &LoadedDoc{DocID: docID, Doc: doc},
	}
	ent.elem = c.ll.PushFront(ent)
	c.cache[keyStr] = ent

	// 超出容量则移除最久未使用的
	if c.maxEntries != 0 && c.ll.Len() > c.maxEntries {
		c.removeOldest()
	}
}

func (c *DocsCache) Delete(key []byte) {
	if ent, ok := c.cache[util.Bytes2String(key)]; ok {
		c.removeElement(ent)
	}
}

func (c *DocsCache) Clear() {
	c.ll = list.New()
	c.cache = make(map[string]*cacheEntry)
}

func (c *DocsCache) DocCount() int {
	return len(c.cache)
}

func (c *DocsCache) removeOldest() {
	if ele := c.ll.Back(); ele != nil {
		c.removeElement(ele.Value.(*cacheEntry))
	}
}

func (c *DocsCache) removeElement(ent *cacheEntry) {
	c.ll.Remove(ent.elem)
	delete(c.cache, ent.key)
}
