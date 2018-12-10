package ravendb

import (
	//"fmt"

	"math"
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
	//fmt.Printf("genericCache.put(): url: %s, changeVector: %s, len(result): %d\n", uri, *i.changeVector, len(i.payload))

	// TODO: probably implement cache eviction
	c.data[uri] = i
}

type HttpCache struct {
	items      *genericCache
	generation atomicInteger
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
		data: map[string]*HttpCacheItem{},
	}
	return &HttpCache{
		items: cache,
	}
}

func (c *HttpCache) GetNumberOfItems() int {
	return c.items.size()
}

func (c *HttpCache) Close() {
	c.items.invalidateAll()
	c.items = nil
}

func (c *HttpCache) set(url string, changeVector *string, result []byte) {
	httpCacheItem := NewHttpCacheItem()
	httpCacheItem.changeVector = changeVector
	httpCacheItem.payload = result
	httpCacheItem.cache = c
	httpCacheItem.generation = c.generation.get()
	c.items.put(url, httpCacheItem)
}

// returns cacheItem, changeVector and response
func (c *HttpCache) get(url string) (*ReleaseCacheItem, *string, []byte) {
	item := c.items.getIfPresent(url)
	if item != nil {
		//fmt.Printf("HttpCache.get(): found url: %s, changeVector: %s, len(payload): %d\n", url, *item.changeVector, len(item.payload))
		return NewReleaseCacheItem(item), item.changeVector, item.payload
	}

	//fmt.Printf("HttpCache.get(): didn't find url: %s\n", url)
	return NewReleaseCacheItem(nil), nil, nil
}

func (c *HttpCache) setNotFound(url string) {
	//fmt.Printf("HttpCache.setNotFound(): url: %s\n", url)
	httpCacheItem := NewHttpCacheItem()
	s := "404 response"
	httpCacheItem.changeVector = &s
	httpCacheItem.cache = c
	httpCacheItem.generation = c.generation.get()

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
	currGen := i.item.generation
	itemGen := i.item.cache.generation.get()
	return currGen != itemGen
}

func (i *ReleaseCacheItem) Close() {
}
