package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cache "github.com/vodolaz095/gin-cache"
	"github.com/vodolaz095/gin-cache/memory"
	rc "github.com/vodolaz095/gin-cache/redis"
)

func main() {
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
	r1.Use(cache.New(memoryCache, cache.CacheByPath(time.Second)))
	// this will be cached with key `/memoryCache/time` and ttl 1 second
	r1.GET("/time", func(c *gin.Context) {
		c.String(http.StatusOK, "Memory cache used! Current time is %s", time.Now().Format(time.Stamp))
	})

	// Redis cache usage example
	r2 := app.Group("/redisCache")
	r2.Use(cache.New(redisCache, func(c *gin.Context) (key string, ttl time.Duration, err error) {
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
