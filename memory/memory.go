package cache

import (
	"sync"
	"time"

	parent "github.com/vodolaz095/gin-cache"
)

// MemoryCache is memory cache storage engine
type MemoryCache struct {
	sync.RWMutex
	items              map[string]parent.Data
	expirationInterval time.Duration
}

// New creates memory cache driver
func New(expirationInterval time.Duration) *MemoryCache {
	cache := MemoryCache{
		items:              make(map[string]parent.Data),
		expirationInterval: expirationInterval,
	}
	if expirationInterval > 0 {
		go cache.startGC()
	}
	return &cache
}

// Save saves item in cache
func (m *MemoryCache) Save(key string, data parent.Data) (err error) {
	m.Lock()
	defer m.Unlock()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	m.items[key] = data
	return nil
}

// Get extracts item from cache
func (m *MemoryCache) Get(key string) (data parent.Data, found bool, err error) {
	m.RLock()
	defer m.RUnlock()
	data, found = m.items[key]
	if !found {
		return
	}
	return
}

// Delete deletes item from cache
func (m *MemoryCache) Delete(key string) (err error) {
	m.Lock()
	defer m.Unlock()
	_, found := m.items[key]
	if found {
		delete(m.items, key)
	}
	return
}

func (m *MemoryCache) startGC() {
	tc := time.NewTicker(m.expirationInterval)
	for t := range tc.C {
		for key, item := range m.items {
			if item.ExpiresAt.After(t) {
				m.Delete(key)
			}
		}
	}
}
