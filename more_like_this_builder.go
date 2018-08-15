package ravendb

var _ IMoreLikeThisOperations = &MoreLikeThisBuilder{}
var _ IMoreLikeThisBuilderForDocumentQuery = &MoreLikeThisBuilder{}
var _ IMoreLikeThisBuilderBase = &MoreLikeThisBuilder{}

type MoreLikeThisBuilder struct {
	moreLikeThis MoreLikeThisBase
}

func NewMoreLikeThisBuilder() *MoreLikeThisBuilder {
	return &MoreLikeThisBuilder{}
}

func (b *MoreLikeThisBuilder) getMoreLikeThis() MoreLikeThisBase {
	return b.moreLikeThis
}

func (b *MoreLikeThisBuilder) usingAnyDocument() IMoreLikeThisOperations {
	b.moreLikeThis = NewMoreLikeThisUsingAnyDocument()
	return b
}

func (b *MoreLikeThisBuilder) usingDocument(documentJson string) IMoreLikeThisOperations {
	b.moreLikeThis = NewMoreLikeThisUsingDocument(documentJson)

	return b
}

func (b *MoreLikeThisBuilder) usingDocumentWithBuilder(builder func(*IFilterDocumentQueryBase)) IMoreLikeThisOperations {
	tmp := NewMoreLikeThisUsingDocumentForDocumentQuery()
	tmp.setForDocumentQuery(builder)
	b.moreLikeThis = tmp
	return b
}

func (b *MoreLikeThisBuilder) WithOptions(options *MoreLikeThisOptions) IMoreLikeThisOperations {
	b.moreLikeThis.SetOptions(options)

	return b
}
