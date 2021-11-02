package gincache

import (
	"github.com/gin-gonic/gin"
	"time"
)

// CacheByPath returns cache key extracting function that uses full request url path
func CacheByPath(duration time.Duration) func(c *gin.Context) (key string, ttl time.Duration, err error) {
	return func(c *gin.Context) (key string, ttl time.Duration, err error) {
		return c.Request.URL.Path, duration, nil
	}
}
