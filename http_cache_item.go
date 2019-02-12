package ravendb

import "time"

type httpCacheItem struct {
	changeVector     *string // TODO: can probably be string
	payload          []byte
	lastServerUpdate time.Time
	generation       int // TODO: should this be atomicInteger?

	cache *httpCache
}

func newHttpCacheItem() *httpCacheItem {
	return &httpCacheItem{
		lastServerUpdate: time.Now(),
	}
}
