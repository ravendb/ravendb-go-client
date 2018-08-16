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

func (o *IndexFieldOptions) GetStorage() FieldStorage {
	return o.storage
}

func (o *IndexFieldOptions) SetStorage(storage FieldStorage) {
	o.storage = storage
}

func (o *IndexFieldOptions) GetIndexing() FieldIndexing {
	return o.indexing
}

func (o *IndexFieldOptions) SetIndexing(indexing FieldIndexing) {
	o.indexing = indexing
}

func (o *IndexFieldOptions) GetTermVector() FieldTermVector {
	return o.termVector
}

func (o *IndexFieldOptions) SetTermVector(termVector FieldTermVector) {
	o.termVector = termVector
}

func (o *IndexFieldOptions) GetSpatial() *SpatialOptions {
	return o.spatial
}

func (o *IndexFieldOptions) SetSpatial(spatial *SpatialOptions) {
	o.spatial = spatial
}

func (o *IndexFieldOptions) GetAnalyzer() string {
	return o.analyzer
}

func (o *IndexFieldOptions) SetAnalyzer(analyzer string) {
	o.analyzer = analyzer
}

func (o *IndexFieldOptions) IsSuggestions() bool {
	return o.suggestions
}

func (o *IndexFieldOptions) SetSuggestions(suggestions bool) {
	o.suggestions = suggestions
}
