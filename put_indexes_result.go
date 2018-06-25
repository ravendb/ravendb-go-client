package ravendb

type PutIndexResult struct {
	IndexName string `json:"IndexName"`
}

func (r *PutIndexResult) getIndexName() string {
	return r.IndexName
}

func (r *PutIndexResult) setIndexName(indexName string) {
	r.IndexName = indexName
}
