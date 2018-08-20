package ravendb

type IMoreLikeThisBuilderForDocumentQuery interface {
	// Note: it's usingDocument() in Java but conflicts with IMoreLikeThisBuilderBase
	UsingDocumentWithBuilder(builder func(*IFilterDocumentQueryBase)) IMoreLikeThisOperations
}
