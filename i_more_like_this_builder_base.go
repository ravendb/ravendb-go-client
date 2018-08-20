package ravendb

type IMoreLikeThisBuilderBase interface {
	UsingAnyDocument() IMoreLikeThisOperations
	UsingDocument(documentJson string) IMoreLikeThisOperations
}
