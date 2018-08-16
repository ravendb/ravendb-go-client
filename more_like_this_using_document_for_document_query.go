package ravendb

var _ MoreLikeThisBase = &MoreLikeThisUsingDocumentForDocumentQuery{}

type MoreLikeThisUsingDocumentForDocumentQuery struct {
	MoreLikeThisCommon

	forDocumentQuery func(*IFilterDocumentQueryBase)
}

func NewMoreLikeThisUsingDocumentForDocumentQuery() *MoreLikeThisUsingDocumentForDocumentQuery {
	return &MoreLikeThisUsingDocumentForDocumentQuery{}
}

func (m *MoreLikeThisUsingDocumentForDocumentQuery) GetForDocumentQuery() func(*IFilterDocumentQueryBase) {
	return m.forDocumentQuery
}

func (m *MoreLikeThisUsingDocumentForDocumentQuery) setForDocumentQuery(forDocumentQuery func(*IFilterDocumentQueryBase)) {
	m.forDocumentQuery = forDocumentQuery
}
