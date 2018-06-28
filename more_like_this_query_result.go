package ravendb

type MoreLikeThisQueryResult struct {
	QueryResultBase
	durationInMs int64
}

func (r *MoreLikeThisQueryResult) getDurationInMs() int64 {
	return r.durationInMs
}

func (r *MoreLikeThisQueryResult) setDurationInMs(durationInMs int64) {
	r.durationInMs = durationInMs
}
