package ravendb

type Query struct {
	collection string
	indexName  string
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) getCollection() string {
	return q.collection
}

func (q *Query) getIndexName() string {
	return q.indexName
}

func Query_index(indexName string) *Query {
	return &Query{
		indexName: indexName,
	}
}

func Query_collection(collectionName string) *Query {
	return &Query{
		collection: collectionName,
	}
}
