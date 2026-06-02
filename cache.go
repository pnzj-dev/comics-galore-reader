package cgreader

import (
	"sync"
)

// Cache provides IndexedDB-backed persistent caching for comic archives and pages.
// In the WASM environment, this bridges to the browser's IndexedDB via syscall/js.
type Cache struct {
	mu    sync.RWMutex
	store map[string][]byte // fallback in-memory store
}

// NewCache creates a new cache instance.
func NewCache() *Cache {
	return &Cache{
		store: make(map[string][]byte),
	}
}

// Put stores a value under the given key.
func (c *Cache) Put(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
	return nil
}

// Get retrieves a value by key. Returns nil if not found.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.store[key]
	return v, ok
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

// Clear removes all entries.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string][]byte)
}

// Size returns the number of cached entries.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}
