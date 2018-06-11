package ravendb

// TODO: implement me

type HttpCache struct {
}

func NewHttpCache() *HttpCache {
	return &HttpCache{}
}

func (c *HttpCache) set(url String, changeVector String, result String) {
}

func (c *HttpCache) setNotFound(url string) {
}

// TODO: implement me
type ReleaseCacheItem struct {
}
