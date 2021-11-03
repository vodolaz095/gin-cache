// Package rcache implements redis cache. This implementation is more suitable for production that memory one, because
// if process restarts, all cached data is persisted in redis, also few webserver processes can share same cache via redis
// database. Unfortunately, redis database should be installed separately.
package rcache
