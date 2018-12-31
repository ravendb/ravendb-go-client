package ravendb

var _ IContent = &IndexQueryContent{}

type IndexQueryContent struct {
	_conventions *DocumentConventions
	_query       *IndexQuery
}

func NewIndexQueryContent(conventions *DocumentConventions, query *IndexQuery) *IndexQueryContent {
	return &IndexQueryContent{
		_conventions: conventions,
		_query:       query,
	}
}

func (q *IndexQueryContent) writeContent() map[string]interface{} {
	return jsonExtensionsWriteIndexQuery(q._conventions, q._query)
}
