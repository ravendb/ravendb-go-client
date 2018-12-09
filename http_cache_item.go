package ravendb

import "time"

type HttpCacheItem struct {
	changeVector     *string // TODO: can probably be string
	payload          string
	lastServerUpdate time.Time
	generation       int // TODO: should this be atomicInteger?

	cache *HttpCache
}

func NewHttpCacheItem() *HttpCacheItem {
	return &HttpCacheItem{
		lastServerUpdate: time.Now(),
	}
}
