package ravendb

type HttpCache struct {
}

func NewHttpCache() *HttpCache {
	return &HttpCache{}
}

func (c *HttpCache) set(url String, changeVector String, result String) {
}
