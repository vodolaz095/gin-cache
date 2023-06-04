GinCache
====================================
Unobtrusive solution to Gin 1.7.2 framework - cache response content in memory or in Redis database.
Module is inspired by this excellent npm package [express-view-cache](https://www.npmjs.com/package/express-view-cache)

[![PkgGoDev](https://pkg.go.dev/badge/github.com/vodolaz095/gin-cache)](https://pkg.go.dev/github.com/vodolaz095/gin-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/vodolaz095/gin-cache)](https://goreportcard.com/report/github.com/vodolaz095/gin-cache)


Advertisement
====================================
You can support development of this module by sending me money directly
https://www.tinkoff.ru/rm/ostroumov.anatoliy2/4HFzm76801/

Why do we need this plugin and how does it work?
====================================

Let's consider we have a GIN application with code like this:

```go

    app := gin.Default()
    // lot of code
    app.GET("/popularPosts", func(c *gin.Context) {
        posts, err := models.GetPopularPosts()
		if err != nil {
			panic(err)
        }
    	c.HTML(http.StatusOK, "popularPosts.tmpl", gin.H{
        	"title": "Popular posts",
			"posts": posts,			
	    })
	})
    // lot of code

```

Function `models.GetPopularPosts` requires a call to database and executed slowly. Also rendering the template 
of posts requires some time. So, maybe we need to cache all this? Ideally, when visitor 
gets the page with url /getPopularPosts we have to give him info right from cache, 
without requests to database, parsing data received, rendering page and other things we need to do 
to give him this page. The most GIN way to do it is to make a separate middleware, that is 
ran before request handler, and returns page from cache (if it is present in cache) or pass 
data to other middlewares, but this caching middleware SAVES rendered response to cache. 
And for future use, the response is taken from CACHE!

Example
==================
This is a complete example of Gin 1.7.4 application which responds with current time:

```go

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
	// function `cache.CacheByPath` returns simple key and ttl extractor function, 
	// that extracts key from full request URI, currently, it will be `/memoryCache/time`
	// and provides cache TTL for duration of 1 second
	r1.GET("/time", func(c *gin.Context) {
		c.String(http.StatusOK, "Memory cache used! Current time is %s", time.Now().Format(time.Stamp))
	})

	// Redis cache usage example
	// We provide our custom analog of `cache.CacheByPath` function, that takes into account
	// request context. 
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


```

Which caching backend implementation to use?
=====================

Module is currently provided with two backends - 
[memory](https://pkg.go.dev/pkg/github.com/vodolaz095/gin-cache/memory/) and [redis](https://pkg.go.dev/pkg/github.com/vodolaz095/gin-cache/redis/).
Both modules satisfies same [Cache interface](https://pkg.go.dev/pkg/github.com/vodolaz095/gin-cache/#Cache), but they
have subtle nuances and differences.

***Memory backend***

Pros:

- simple to configure
- no dependency on any 3rd party services
- no extra network/socket descriptor usage

Cons:

- if we restart process, cache is purged
- cache cannot be shared between processes
- extra RAM consumption by application

***Redis backend***

Pros:

- cache persist, if application restarts
- cache can be shared between processes
- application consumes less ram

Cons:

- separate redis server is required
- at least one extra network/socket descriptor is consumed 


Testing code 
======================

```shell

$ make deps
$ make lint
$ make check

```

Start example
=====================

```shell

make start

```

And open [http://localhost:3000](http://localhost:3000) in browser.


LICENSE
=====================

The MIT License (MIT)

Copyright (c) 2021 Ostroumov Anatolij <ostroumov095 at gmail dot com>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

