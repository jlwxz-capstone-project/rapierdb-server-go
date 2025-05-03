package storage_engine

import (
	"container/list"
	"sync"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

const DEFAULT_CACHE_SIZE = 1024 // Default cache size

// LoadedDoc holds the document ID and the Loro document instance.
type LoadedDoc struct {
	DocID string
	Doc   *loro.LoroDoc
}

// cacheEntry represents an entry in the LRU cache.
type cacheEntry struct {
	key  string        // The key used to store the entry in the map.
	doc  *LoadedDoc    // Pointer to the loaded document.
	elem *list.Element // Pointer to the element in the linked list for LRU tracking.
}

// DocsCache implements an LRU cache for Loro documents.
// It uses a map for quick lookups and a doubly linked list to track usage order.
// Access to the cache is protected by a RWMutex.
type DocsCache struct {
	mu         sync.RWMutex           // Mutex to protect concurrent access to the cache.
	maxEntries int                    // Maximum number of entries the cache can hold. 0 means no limit.
	ll         *list.List             // Doubly linked list for LRU tracking. Front is most recently used.
	cache      map[string]*cacheEntry // Map storing cache entries, keyed by the document key string.
}

// NewEmptyDocsCache creates a new DocsCache with the default size.
func NewEmptyDocsCache() *DocsCache {
	return &DocsCache{
		maxEntries: DEFAULT_CACHE_SIZE,
		ll:         list.New(),
		cache:      make(map[string]*cacheEntry),
	}
}

// Get retrieves a document from the cache by its key.
// If the document is found, it's marked as recently used and returned.
// Returns the document and true if found, otherwise nil and false.
// This operation requires a write lock because it modifies the LRU list order.
func (c *DocsCache) Get(key []byte) (*loro.LoroDoc, bool) {
	c.mu.Lock() // Acquire write lock as we might modify the list
	defer c.mu.Unlock()

	keyStr := util.Bytes2String(key)
	if ent, ok := c.cache[keyStr]; ok {
		c.ll.MoveToFront(ent.elem) // Mark as recently used
		return ent.doc.Doc, true
	}
	return nil, false
}

// Set adds or updates a document in the cache.
// If the document already exists, its value is updated, and it's marked as recently used.
// If the cache exceeds its maximum size after adding, the least recently used item is removed.
// This operation requires a write lock.
func (c *DocsCache) Set(key []byte, docID string, doc *loro.LoroDoc) {
	c.mu.Lock() // Acquire write lock
	defer c.mu.Unlock()

	keyStr := util.Bytes2String(key)
	if ent, ok := c.cache[keyStr]; ok {
		// Entry exists, update it and move to front
		c.ll.MoveToFront(ent.elem)
		ent.doc = &LoadedDoc{DocID: docID, Doc: doc}
		return
	}

	// Entry doesn't exist, add a new one
	ent := &cacheEntry{
		key: keyStr,
		doc: &LoadedDoc{DocID: docID, Doc: doc},
	}
	ent.elem = c.ll.PushFront(ent) // Add to front (most recently used)
	c.cache[keyStr] = ent

	// Check if eviction is needed
	if c.maxEntries != 0 && c.ll.Len() > c.maxEntries {
		c.removeOldest() // removeOldest is called while lock is held
	}
}

// Delete removes a document from the cache by its key.
// This operation requires a write lock.
func (c *DocsCache) Delete(key []byte) {
	c.mu.Lock() // Acquire write lock
	defer c.mu.Unlock()

	keyStr := util.Bytes2String(key)
	if ent, ok := c.cache[keyStr]; ok {
		c.removeElement(ent) // removeElement is called while lock is held
	}
}

// Clear removes all entries from the cache.
// This operation requires a write lock.
func (c *DocsCache) Clear() {
	c.mu.Lock() // Acquire write lock
	defer c.mu.Unlock()

	c.ll = list.New()
	c.cache = make(map[string]*cacheEntry)
}

// DocCount returns the number of documents currently in the cache.
// This operation uses a read lock.
func (c *DocsCache) DocCount() int {
	c.mu.RLock() // Acquire read lock
	defer c.mu.RUnlock()
	return len(c.cache)
}

// removeOldest removes the least recently used item from the cache.
// Assumes the write lock is already held by the caller.
func (c *DocsCache) removeOldest() {
	if ele := c.ll.Back(); ele != nil {
		c.removeElement(ele.Value.(*cacheEntry))
	}
}

// removeElement removes a specific cache entry from the list and map.
// Assumes the write lock is already held by the caller.
func (c *DocsCache) removeElement(ent *cacheEntry) {
	c.ll.Remove(ent.elem)
	delete(c.cache, ent.key)
}
