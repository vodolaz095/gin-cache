package gincache

import (
	"time"
)

// Data depicts cached response
type Data struct {
	Key         string
	Body        []byte
	Status      int
	ContentType string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// Cache is interface to be used with different caching backends
type Cache interface {
	Save(key string, data Data) (err error)
	Get(key string) (data Data, found bool, err error)
	Delete(key string) (err error)
}
