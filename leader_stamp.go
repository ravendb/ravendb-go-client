package ravendb

// LeaderStamp describes leader stamp
type LeaderStamp struct {
	Index        int64 `json:"Index"`
	Term         int64 `json:"Term"`
	LeadersTicks int64 `json:"LeadersTicks"`
}
