package ravendb

type IndexFieldOptions struct {
	storage     FieldStorage
	indexing    FieldIndexing
	termVector  FieldTermVector
	spatial     *SpatialOptions
	analyzer    string
	suggestions bool
}

func NewIndexFieldOptions() *IndexFieldOptions {
	return &IndexFieldOptions{}
}

func (o *IndexFieldOptions) getStorage() FieldStorage {
	return o.storage
}

func (o *IndexFieldOptions) setStorage(storage FieldStorage) {
	o.storage = storage
}

func (o *IndexFieldOptions) getIndexing() FieldIndexing {
	return o.indexing
}

func (o *IndexFieldOptions) setIndexing(indexing FieldIndexing) {
	o.indexing = indexing
}

func (o *IndexFieldOptions) getTermVector() FieldTermVector {
	return o.termVector
}

func (o *IndexFieldOptions) setTermVector(termVector FieldTermVector) {
	o.termVector = termVector
}

func (o *IndexFieldOptions) getSpatial() *SpatialOptions {
	return o.spatial
}

func (o *IndexFieldOptions) setSpatial(spatial *SpatialOptions) {
	o.spatial = spatial
}

func (o *IndexFieldOptions) getAnalyzer() string {
	return o.analyzer
}

func (o *IndexFieldOptions) setAnalyzer(analyzer string) {
	o.analyzer = analyzer
}

func (o *IndexFieldOptions) isSuggestions() bool {
	return o.suggestions
}

func (o *IndexFieldOptions) setSuggestions(suggestions bool) {
	o.suggestions = suggestions
}
