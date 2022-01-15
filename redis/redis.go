package rcache

import (
	"fmt"
	"github.com/go-redis/redis"
	"net/url"
	"strconv"
	"strings"
	"time"

	parent "github.com/vodolaz095/gin-cache"
)

// DefaultConnectionString is a usual way to connect to redis running on 127.0.0.1:6379 without password authentication, and we use database 0
const DefaultConnectionString = "redis://127.0.0.1:6379/0"

// Cache is redis storage engine
type Cache struct {
	prefix string
	client *redis.Client
}

// ParseConnectionString parses connection string to generate redis connection options
func ParseConnectionString(connectionString string) (options redis.Options, err error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return
	}
	if u.Scheme != "redis" {
		err = fmt.Errorf("unknown protocol %s - only \"redis\" allowed", u.Scheme)
		return
	}
	options.Addr = u.Host
	if u.User != nil {
		pwd, present := u.User.Password()
		if present {
			options.Password = pwd
		}
	}
	if u.Path != "" {
		dbTrimmed := strings.TrimPrefix(u.Path, "/")
		dbn, errP := strconv.ParseUint(dbTrimmed, 10, 64)
		if errP != nil {
			err = fmt.Errorf("%s - while parsing redis database number >>>%s<<< as positive integer, like 4 in connection string redis://127.0.0.1:6379/4",
				errP,
				dbTrimmed,
			)
			return
		}
		options.DB = int(dbn)
	}
	return
}

// New creates new redis caching driver
func New(redisConnectionString, prefix string) (rc *Cache, err error) {
	rc = &Cache{prefix: prefix}
	opts, err := ParseConnectionString(redisConnectionString)
	if err != nil {
		return
	}
	rc.client = redis.NewClient(&opts)
	pong, err := rc.client.Ping().Result()
	if err != nil {
		return
	}
	if pong != "PONG" {
		err = fmt.Errorf("wrong ping response")
		return
	}
	return rc, nil
}

// Save saves item in cache
func (rc *Cache) Save(key string, data parent.Data) (err error) {
	prefixedKey := fmt.Sprintf("%s%s", rc.prefix, key)
	_, err = rc.client.HMSet(prefixedKey, map[string]interface{}{
		"key":         key,
		"body":        string(data.Body),
		"status":      fmt.Sprintf("%v", data.Status),
		"contentType": data.ContentType,
		"createdAt":   data.CreatedAt.Format(time.RFC1123),
		"expiresAt":   data.ExpiresAt.Format(time.RFC1123),
	}).Result()
	if err != nil {
		return
	}
	_, err = rc.client.ExpireAt(prefixedKey, data.ExpiresAt).Result()
	return
}

// Get extracts item from cache
func (rc *Cache) Get(key string) (data parent.Data, found bool, err error) {
	key = fmt.Sprintf("%s%s", rc.prefix, key)
	data = parent.Data{}
	raw, err := rc.client.HGetAll(key).Result()
	if err != nil {
		return
	}
	if len(raw) == 0 {
		return data, false, nil
	}
	found = true
	data.Key = raw["key"]
	data.ContentType = raw["contentType"]
	status, err := strconv.ParseInt(raw["status"], 10, 16)
	if err != nil {
		return
	}
	data.Status = int(status)
	data.Body = []byte(raw["body"])
	createdAt, err := time.Parse(time.RFC1123, raw["createdAt"])
	if err != nil {
		return
	}
	data.CreatedAt = createdAt
	expiresAt, err := time.Parse(time.RFC1123, raw["expiresAt"])
	if err != nil {
		return
	}
	data.ExpiresAt = expiresAt
	return
}

// Delete deletes item from cache
func (rc *Cache) Delete(key string) (err error) {
	key = fmt.Sprintf("%s%s", rc.prefix, key)
	return rc.client.Del(key).Err()
}
