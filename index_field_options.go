package ravendb

type IndexFieldOptions struct {
	Storage     FieldStorage    `json:"Storage,omitempty"`
	Indexing    FieldIndexing   `json:"Indexing,omitempty"`
	TermVector  FieldTermVector `json:"TermVector,omitempty"`
	Spatial     *SpatialOptions `json:"Spatial"`
	Analyzer    string          `json:"Analyzer,omitempty"`
	Suggestions bool            `json:"Suggestions"`
}

func NewIndexFieldOptions() *IndexFieldOptions {
	return &IndexFieldOptions{}
}
