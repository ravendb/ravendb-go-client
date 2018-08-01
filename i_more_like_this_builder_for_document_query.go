package ravendb

type IMoreLikeThisBuilderForDocumentQuery interface {
	// Note: it's usingDocument() in Java but conflicts with IMoreLikeThisBuilderBase
	usingDocumentWithBuilder(builder func(*IFilterDocumentQueryBase)) IMoreLikeThisOperations
}
