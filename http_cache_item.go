package ravendb

import "time"

type HttpCacheItem struct {
	changeVector     string
	payload          string
	lastServerUpdate time.Time
	generation       int

	cache *HttpCache
}

func NewHttpCacheItem() *HttpCacheItem {
	return &HttpCacheItem{
		lastServerUpdate: time.Now(),
	}
}
