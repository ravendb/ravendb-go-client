package ravendb

type MoreLikeThisBase interface {
	GetOptions() *MoreLikeThisOptions
	SetOptions(options *MoreLikeThisOptions)
}

type MoreLikeThisCommon struct {
	options *MoreLikeThisOptions
}

func (c *MoreLikeThisCommon) GetOptions() *MoreLikeThisOptions {
	return c.options
}

func (c *MoreLikeThisCommon) SetOptions(options *MoreLikeThisOptions) {
	c.options = options
}
