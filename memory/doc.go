// Package memory implements simple in memory cache, that is automatically purged of expired entries to prevent memleaks.
// This implementation cannot be considered reliable and production ready, because if process restarts,
// all cached data is lost. Also cached data cannot be shared between different instances.
package memory
