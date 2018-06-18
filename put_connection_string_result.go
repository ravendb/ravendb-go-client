package ravendb

type PutConnectionStringResult struct {
	Etag int `json:"ETag"`
}

func (r *PutConnectionStringResult) getEtag() int {
	return r.Etag
}

func (r *PutConnectionStringResult) setEtag(etag int) {
	r.Etag = etag
}
