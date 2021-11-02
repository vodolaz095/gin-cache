package memory

import (
	"sync"
	"time"

	parent "github.com/vodolaz095/gin-cache"
)

// Cache is memory cache storage engine
type Cache struct {
	sync.RWMutex
	items              map[string]parent.Data
	expirationInterval time.Duration
}

// New creates memory cache driver
func New(expirationInterval time.Duration) *Cache {
	cache := Cache{
		items:              make(map[string]parent.Data),
		expirationInterval: expirationInterval,
	}
	if expirationInterval > 0 {
		go cache.startGC()
	}
	return &cache
}

// Save saves item in cache
func (m *Cache) Save(key string, data parent.Data) (err error) {
	m.Lock()
	defer m.Unlock()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	data.Key = key
	m.items[key] = data
	return nil
}

// Get extracts item from cache
func (m *Cache) Get(key string) (data parent.Data, found bool, err error) {
	m.RLock()
	defer m.RUnlock()
	data, found = m.items[key]
	if !found {
		return
	}
	return
}

// Delete deletes item from cache
func (m *Cache) Delete(key string) (err error) {
	m.Lock()
	defer m.Unlock()
	_, found := m.items[key]
	if found {
		delete(m.items, key)
	}
	return
}

func (m *Cache) startGC() {
	tc := time.NewTicker(m.expirationInterval)
	for t := range tc.C {
		for key, item := range m.items {
			if item.ExpiresAt.Before(t) {
				m.Delete(key)
			}
		}
	}
}
