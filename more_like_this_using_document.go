package ravendb

var _ MoreLikeThisBase = &MoreLikeThisUsingDocument{}

// MoreLikeThisUsingDocument represents more like this with a document
type MoreLikeThisUsingDocument struct {
	MoreLikeThisCommon

	documentJSON string
}
