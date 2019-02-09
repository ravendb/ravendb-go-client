package ravendb

// MoreLikeThisQueryResult describes result of "more like this" operation
type MoreLikeThisQueryResult struct {
	queryResultBase
	DurationInMs int64 `json:"DurationInMs"`
}
