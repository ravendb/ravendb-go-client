package ravendb

type IMoreLikeThisBuilderBase interface {
	usingAnyDocument() IMoreLikeThisOperations
	usingDocument(documentJson string) IMoreLikeThisOperations
}
