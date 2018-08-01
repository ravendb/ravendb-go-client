package ravendb

var _ MoreLikeThisBase = &MoreLikeThisUsingDocument{}

type MoreLikeThisUsingDocument struct {
	MoreLikeThisCommon

	documentJson string
}

func NewMoreLikeThisUsingDocument(documentJson string) *MoreLikeThisUsingDocument {
	return &MoreLikeThisUsingDocument{
		documentJson: documentJson,
	}
}

func (m *MoreLikeThisUsingDocument) getDocumentJson() string {
	return m.documentJson
}

func (m *MoreLikeThisUsingDocument) setDocumentJson(documentJson string) {
	m.documentJson = documentJson
}
