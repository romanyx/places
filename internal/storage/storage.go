package storage

import "errors"

// ErrCacheNotFound returns when there is no cache
// in storage for a given query.
var ErrCacheNotFound = errors.New("cache not found")
