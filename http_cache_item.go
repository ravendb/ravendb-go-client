package ravendb

import "time"

type HttpCacheItem struct {
	changeVector     *string // TODO: can probably be string
	payload          string
	lastServerUpdate time.Time
	generation       int32

	cache *HttpCache
}

func NewHttpCacheItem() *HttpCacheItem {
	return &HttpCacheItem{
		lastServerUpdate: time.Now(),
	}
}
