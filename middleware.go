package gincache

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Good read
// https://github.com/gin-gonic/gin/issues/1363#issuecomment-577722498

type sniffer struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (s sniffer) Write(b []byte) (int, error) {
	s.body.Write(b)
	return s.ResponseWriter.Write(b)
}

func (s sniffer) WriteString(payload string) (n int, err error) {
	s.body.WriteString(payload)
	return s.ResponseWriter.WriteString(payload)
}

// New creates new caching middleware with cache and extractor function provided
func New(
	cache Cache,
	keyExtractor func(c *gin.Context) (key string, ttl time.Duration, err error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// only get request responses can be cached
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		key, ttl, err := keyExtractor(c)
		if err != nil {
			panic(err)
		}
		data, found, err := cache.Get(key)
		if err != nil {
			panic(err)
		}
		if found {
			c.Header("Last-Modified", data.CreatedAt.Format(time.RFC1123))
			c.Header("Expires", data.ExpiresAt.Format(time.RFC1123))
			c.Data(data.Status, data.ContentType, data.Body)
			c.Abort()
			return
		}
		now := time.Now()
		c.Header("Last-Modified", now.Format(time.RFC1123))
		c.Header("Expires", now.Add(ttl).Format(time.RFC1123))
		s := &sniffer{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = s
		c.Next()
		// saving sniffed body
		newDataToBeSaved := Data{
			Key:         key,
			Body:        s.body.Bytes(),
			Status:      c.Writer.Status(),
			ContentType: c.Writer.Header().Get("Content-Type"),
			CreatedAt:   time.Now(),
			ExpiresAt:   time.Now().Add(ttl),
		}
		err = cache.Save(key, newDataToBeSaved)
		if err != nil {
			panic(err)
		}
	}
}
