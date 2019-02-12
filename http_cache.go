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
	weighter      func(string, *httpCacheItem) int

	data map[string]*httpCacheItem
}

func (c *genericCache) size() int {
	return len(c.data)
}

func (c *genericCache) invalidateAll() {
	c.data = nil
}

func (c *genericCache) getIfPresent(uri string) *httpCacheItem {
	return c.data[uri]
}

func (c *genericCache) put(uri string, i *httpCacheItem) {
	//fmt.Printf("genericCache.put(): url: %s, changeVector: %s, len(result): %d\n", uri, *i.changeVector, len(i.payload))

	// TODO: probably implement cache eviction
	c.data[uri] = i
}

type httpCache struct {
	items      *genericCache
	generation atomicInteger
}

func newHttpCache(size int) *httpCache {
	if size == 0 {
		size = 1 * 1024 * 1024 // TODO: check what is default size of com.google.common.cache.Cache is
	}
	cache := &genericCache{
		softValues:    true,
		maximumWeight: size,
		weighter: func(k string, v *httpCacheItem) int {
			return len(v.payload) + 20
		},
		data: map[string]*httpCacheItem{},
	}
	return &httpCache{
		items: cache,
	}
}

func (c *httpCache) GetNumberOfItems() int {
	return c.items.size()
}

func (c *httpCache) close() {
	c.items.invalidateAll()
	c.items = nil
}

func (c *httpCache) set(url string, changeVector *string, result []byte) {
	httpCacheItem := newHttpCacheItem()
	httpCacheItem.changeVector = changeVector
	httpCacheItem.payload = result
	httpCacheItem.cache = c
	httpCacheItem.generation = c.generation.get()
	c.items.put(url, httpCacheItem)
}

// returns cacheItem, changeVector and response
func (c *httpCache) get(url string) (*releaseCacheItem, *string, []byte) {
	item := c.items.getIfPresent(url)
	if item != nil {
		//fmt.Printf("HttpCache.get(): found url: %s, changeVector: %s, len(payload): %d\n", url, *item.changeVector, len(item.payload))
		return newReleaseCacheItem(item), item.changeVector, item.payload
	}

	//fmt.Printf("HttpCache.get(): didn't find url: %s\n", url)
	return newReleaseCacheItem(nil), nil, nil
}

func (c *httpCache) setNotFound(url string) {
	//fmt.Printf("HttpCache.setNotFound(): url: %s\n", url)
	httpCacheItem := newHttpCacheItem()
	s := "404 response"
	httpCacheItem.changeVector = &s
	httpCacheItem.cache = c
	httpCacheItem.generation = c.generation.get()

	c.items.put(url, httpCacheItem)
}

type releaseCacheItem struct {
	item *httpCacheItem
}

func newReleaseCacheItem(item *httpCacheItem) *releaseCacheItem {
	return &releaseCacheItem{
		item: item,
	}
}

func (i *releaseCacheItem) notModified() {
	if i.item != nil {
		i.item.lastServerUpdate = time.Now()
	}
}

func (i *releaseCacheItem) getAge() time.Duration {
	if i.item == nil {
		return time.Duration(math.MaxInt64)
	}
	return time.Since(i.item.lastServerUpdate)
}

func (i *releaseCacheItem) getMightHaveBeenModified() bool {
	currGen := i.item.generation
	itemGen := i.item.cache.generation.get()
	return currGen != itemGen
}

func (i *releaseCacheItem) close() {
	// no-op
}
