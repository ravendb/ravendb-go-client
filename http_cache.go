package ravendb

// TODO: implement me

type HttpCache struct {
}

func NewHttpCache() *HttpCache {
	return &HttpCache{}
}

func (c *HttpCache) set(url string, changeVector string, result string) {
}

func (c *HttpCache) setNotFound(url string) {
}

// TODO: implement me
type ReleaseCacheItem struct {
}
