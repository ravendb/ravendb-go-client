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

func (b *MoreLikeThisBuilder) GetMoreLikeThis() MoreLikeThisBase {
	return b.moreLikeThis
}

func (b *MoreLikeThisBuilder) UsingAnyDocument() IMoreLikeThisOperations {
	b.moreLikeThis = NewMoreLikeThisUsingAnyDocument()
	return b
}

func (b *MoreLikeThisBuilder) UsingDocument(documentJSON string) IMoreLikeThisOperations {
	b.moreLikeThis = &MoreLikeThisUsingDocument{
		documentJSON: documentJSON,
	}

	return b
}

func (b *MoreLikeThisBuilder) UsingDocumentWithBuilder(builder func(*IFilterDocumentQueryBase)) IMoreLikeThisOperations {
	tmp := NewMoreLikeThisUsingDocumentForDocumentQuery()
	tmp.setForDocumentQuery(builder)
	b.moreLikeThis = tmp
	return b
}

func (b *MoreLikeThisBuilder) WithOptions(options *MoreLikeThisOptions) IMoreLikeThisOperations {
	b.moreLikeThis.SetOptions(options)

	return b
}
