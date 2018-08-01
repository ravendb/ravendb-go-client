package ravendb

var _ MoreLikeThisBase = &MoreLikeThisUsingAnyDocument{}

type MoreLikeThisUsingAnyDocument struct {
	MoreLikeThisCommon
}

func NewMoreLikeThisUsingAnyDocument() *MoreLikeThisUsingAnyDocument {
	return &MoreLikeThisUsingAnyDocument{}
}
