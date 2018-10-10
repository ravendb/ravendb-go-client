package ravendb

import (
	"math"
	"sync/atomic"
	"time"
)

// equivalent of com.google.common.cache.Cache, specialized for String -> HttpCacheItem mapping
// TODO: match semantics
type genericCache struct {
	softValues    bool
	maximumWeight int
	weighter      func(string, *HttpCacheItem) int

	data map[string]*HttpCacheItem
}

func (c *genericCache) size() int {
	return len(c.data)
}

func (c *genericCache) invalidateAll() {
	c.data = nil
}

func (c *genericCache) getIfPresent(uri string) *HttpCacheItem {
	return c.data[uri]
}

func (c *genericCache) put(uri string, i *HttpCacheItem) {
	// TODO: probably implement cache eviction
	c.data[uri] = i
}

type HttpCache struct {
	items      *genericCache
	generation int32 // atomic
}

func NewHttpCache(size int) *HttpCache {
	if size == 0 {
		size = 1 * 1024 * 1024 // TODO: check what is default size of com.google.common.cache.Cache is
	}
	cache := &genericCache{
		softValues:    true,
		maximumWeight: size,
		weighter: func(k string, v *HttpCacheItem) int {
			return len(v.payload) + 20
		},
	}
	return &HttpCache{
		items: cache,
	}
}

func (c *HttpCache) getNumberOfItems() int {
	return c.items.size()
}

func (c *HttpCache) Close() {
	c.items.invalidateAll()
	c.items = nil
}

func (c *HttpCache) set(url string, changeVector *string, result string) {
	httpCacheItem := NewHttpCacheItem()
	httpCacheItem.changeVector = changeVector
	httpCacheItem.payload = result
	httpCacheItem.cache = c
	gen := atomic.LoadInt32(&c.generation)
	httpCacheItem.generation = gen
	c.items.put(url, httpCacheItem)
}

// returns changeVector and response
func (c *HttpCache) get(url string) (ci *ReleaseCacheItem, changeVector *string, response string) {
	item := c.items.getIfPresent(url)
	if item != nil {
		changeVector = item.changeVector
		response = item.payload

		ci = NewReleaseCacheItem(item)
		return
	}

	changeVector = nil
	response = ""
	ci = NewReleaseCacheItem(nil)
	return
}

func (c *HttpCache) setNotFound(url string) {
	httpCacheItem := NewHttpCacheItem()
	s := "404 response"
	httpCacheItem.changeVector = &s
	httpCacheItem.cache = c
	gen := atomic.LoadInt32(&c.generation)
	httpCacheItem.generation = gen

	c.items.put(url, httpCacheItem)
}

type ReleaseCacheItem struct {
	item *HttpCacheItem
}

func NewReleaseCacheItem(item *HttpCacheItem) *ReleaseCacheItem {
	return &ReleaseCacheItem{
		item: item,
	}
}

func (i *ReleaseCacheItem) notModified() {
	if i.item != nil {
		i.item.lastServerUpdate = time.Now()
	}
}

func (i *ReleaseCacheItem) getAge() time.Duration {
	if i.item == nil {
		return time.Duration(math.MaxInt64)
	}
	return time.Since(i.item.lastServerUpdate)
}

func (i *ReleaseCacheItem) getMightHaveBeenModified() bool {
	currGen := atomic.LoadInt32(&i.item.generation)
	itemGen := atomic.LoadInt32(&i.item.cache.generation)
	return currGen == itemGen
}

func (i *ReleaseCacheItem) Close() {
}
