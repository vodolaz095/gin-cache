package gincache

import (
	"fmt"
	"github.com/vodolaz095/gin-cache/memory"
	rc "github.com/vodolaz095/gin-cache/redis"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

var testApp *gin.Engine

type testCacher struct {
	sync.RWMutex
	items map[string]Data
	log   func(format string, args ...interface{})
}

// Save saves item in cache
func (m *testCacher) Save(key string, data Data) (err error) {
	fmt.Printf("TestCache: Saving %s with status %v and body %s\n", key, data.Status, string(data.Body))
	m.Lock()
	defer m.Unlock()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	m.items[key] = data
	return nil
}

// Get extracts item from cache
func (m *testCacher) Get(key string) (data Data, found bool, err error) {
	fmt.Printf("TestCache: Extracting key %s...\n", key)
	m.RLock()
	defer m.RUnlock()
	data, found = m.items[key]
	if !found {
		fmt.Printf("TestCache: Key %s not found\n", key)
		return
	}
	fmt.Printf("TestCache: Key %s extracted with status %v and body %s\n", key, data.Status, string(data.Body))
	return
}

// Delete deletes item from cache
func (m *testCacher) Delete(key string) (err error) {
	m.Lock()
	defer m.Unlock()
	_, found := m.items[key]
	if found {
		delete(m.items, key)
	}
	return
}

var testCache *testCacher

func TestPrepare(t *testing.T) {
	t.Logf("Preparing test gin app...")
	testApp = gin.New()
	testCache = &testCacher{items: make(map[string]Data)}
	cacherMiddleware := New(testCache, CacheByPath(time.Second))
	testApp.Use(cacherMiddleware)
	testApp.NoRoute(func(c *gin.Context) {
		c.String(http.StatusTeapot, "Current time is %s", time.Now().Format(time.Stamp))
	})
}

var testFreshBody string
var testCacheEntryCreatedAt string
var testCacheEntryExpiresAt string

func TestGetFresh(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://russian.rt.com/time",
		nil,
	) // GIN engine should ignore HOSTNAME in header, so its ok if i provide it here
	w := httptest.NewRecorder()
	testApp.ServeHTTP(w, req)
	resp := w.Result()
	t.Logf("Status: %s", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	testFreshBody = string(body)
	testCacheEntryCreatedAt = resp.Header.Get("Last-Modified")
	testCacheEntryExpiresAt = resp.Header.Get("Expires")

	t.Logf("Body is %s", testFreshBody)
	t.Logf("Content type is %s", resp.Header.Get("Content-Type"))
	t.Logf("Last-Modified is %s", resp.Header.Get("Last-Modified"))
	t.Logf("Expires is %s", resp.Header.Get("Expires"))
}

func TestEnsureKeyCreatedProperly(t *testing.T) {
	for k, v := range testCache.items {
		t.Logf("Item %s found in cache with body %s", k, string(v.Body))
	}
	if len(testCache.items) != 1 {
		t.Error("no items in cache")
	}
	data, found := testCache.items["/time"]
	if !found {
		t.Error("data not found!")
		return
	}
	t.Logf("Key: %s", data.Key)
	t.Logf("Body: %s", string(data.Body))
	t.Logf("Status: %v", data.Status)
	t.Logf("Content Type: %v", data.ContentType)
}

func TestCacheGet(t *testing.T) {
	data, found, err := testCache.Get("/time")
	if err != nil {
		t.Errorf("%s : while extracting key /time", err)
	}
	if !found {
		t.Error("key not found?")
	}
	t.Logf("Key: %s", data.Key)
	t.Logf("Body: %s", string(data.Body))
	t.Logf("Status: %v", data.Status)
	t.Logf("Content Type: %v", data.ContentType)
}

func TestGetCached(t *testing.T) {
	time.Sleep(time.Second)
	req := httptest.NewRequest(
		"GET",
		"http://russian.rt.com/time",
		nil,
	) // GIN engine should ignore HOSTNAME in header, so its ok if i provide it here
	w := httptest.NewRecorder()
	testApp.ServeHTTP(w, req)
	resp := w.Result()
	t.Logf("Status: %s", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	bodyAsString := string(body)
	t.Logf("Body is %s", bodyAsString)
	if bodyAsString != testFreshBody {
		t.Error("cache returned wrong content")
	}
	if testCacheEntryCreatedAt != resp.Header.Get("Last-Modified") {
		t.Logf("Expected: %s", testCacheEntryCreatedAt)
		t.Logf("Received: %s", resp.Header.Get("Last-Modified"))
		t.Error("cache returned wrong Last-Modified")
	}
	if testCacheEntryExpiresAt != resp.Header.Get("Expires") {
		t.Logf("Expected: %s", testCacheEntryExpiresAt)
		t.Logf("Received: %s", resp.Header.Get("Expires"))
		t.Error("cache returned wrong Expires")
	}
}

func TestPostCached(t *testing.T) {
	time.Sleep(time.Second)
	req := httptest.NewRequest(
		"POST",
		"http://russian.rt.com/time",
		nil,
	) // GIN engine should ignore HOSTNAME in header, so its ok if i provide it here
	w := httptest.NewRecorder()
	testApp.ServeHTTP(w, req)
	resp := w.Result()
	t.Logf("Status: %s", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	bodyAsString := string(body)
	t.Logf("Body is %s", bodyAsString)
	if bodyAsString == testFreshBody {
		t.Error("cache should be bypassed for POST")
	}
}

func Example() {
	var err error
	app := gin.Default()
	memoryCache := memory.New(5 * time.Second)
	redisCache, err := rc.New(rc.DefaultConnectionString, "redisCacheExamplePrefix")
	if err != nil {
		log.Fatalf("%s : while connecting to redis at %s", err, rc.DefaultConnectionString)
	}
	app.Use(func(c *gin.Context) {
		c.Header("Refresh", "1")
		c.Next()
	})

	// Memory cache usage example
	r1 := app.Group("/memoryCache")
	r1.Use(New(memoryCache, CacheByPath(time.Second)))
	// this will be cached with key `/memoryCache/time` and ttl 1 second
	r1.GET("/time", func(c *gin.Context) {
		c.String(http.StatusOK, "Memory cache used! Current time is %s", time.Now().Format(time.Stamp))
	})

	// Redis cache usage example
	r2 := app.Group("/redisCache")
	r2.Use(New(redisCache, func(c *gin.Context) (key string, ttl time.Duration, err error) {
		user, authorised := c.Get("user")
		// if there is no authorized user, we cache data for 1 minute, using customers IP as cache key
		if !authorised {
			return c.ClientIP(), time.Minute, nil
		}
		// if user is authorized, we cache data for 15 second,
		// using string representation of user parameter as cache key
		return fmt.Sprint(user), 15 * time.Second, nil
	}))
	r2.GET("/time", func(c *gin.Context) {
		c.String(http.StatusOK, "Redis Cache used! Current time is %s", time.Now().Format(time.Stamp))
	})

	app.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusOK)
		fmt.Fprintln(c.Writer, "<html><body>")
		fmt.Fprintln(c.Writer, " <p><a href=\"/memoryCache/time\">Test memory cache</p>")
		fmt.Fprintln(c.Writer, " <p><a href=\"/redisCache/time\">Test redis cache</p>")
		fmt.Fprintln(c.Writer, "</body></html>")
		c.Abort()
	})

	err = app.Run("127.0.0.1:3000")
	if err != nil {
		log.Fatalf("%s : while starting app", err)
	}
}
