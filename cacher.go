package gincache

import (
	"context"
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

// Cache is interface to be used with different caching backends. Currently, `memory` and `redis` backends are provided
type Cache interface {
	Save(ctx context.Context, key string, data Data) (err error)
	Get(ctx context.Context, key string) (data Data, found bool, err error)
	Delete(ctx context.Context, key string) (err error)
}
