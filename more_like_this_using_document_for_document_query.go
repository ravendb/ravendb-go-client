package ravendb

var _ MoreLikeThisBase = &MoreLikeThisUsingDocumentForDocumentQuery{}

type MoreLikeThisUsingDocumentForDocumentQuery struct {
	MoreLikeThisCommon

	forDocumentQuery func(*DocumentQuery)
}

func NewMoreLikeThisUsingDocumentForDocumentQuery() *MoreLikeThisUsingDocumentForDocumentQuery {
	return &MoreLikeThisUsingDocumentForDocumentQuery{}
}

func (m *MoreLikeThisUsingDocumentForDocumentQuery) GetForDocumentQuery() func(*DocumentQuery) {
	return m.forDocumentQuery
}

func (m *MoreLikeThisUsingDocumentForDocumentQuery) setForDocumentQuery(forDocumentQuery func(*DocumentQuery)) {
	m.forDocumentQuery = forDocumentQuery
}
