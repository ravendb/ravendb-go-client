package ravendb

type MoreLikeThisBase interface {
	getOptions() *MoreLikeThisOptions
	setOptions(options *MoreLikeThisOptions)
}

type MoreLikeThisCommon struct {
	options *MoreLikeThisOptions
}

func (c *MoreLikeThisCommon) getOptions() *MoreLikeThisOptions {
	return c.options
}

func (c *MoreLikeThisCommon) setOptions(options *MoreLikeThisOptions) {
	c.options = options
}
