package ravendb

type PutIndexResult struct {
	// Note: don't know how Java does it, but this is sent
	// as Index in JSON responses from the server
	IndexName string `json:"Index"`
}

func (r *PutIndexResult) getIndexName() string {
	return r.IndexName
}

func (r *PutIndexResult) setIndexName(indexName string) {
	r.IndexName = indexName
}
