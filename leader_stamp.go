package ravendb

// LeaderStamp describes leader stamp
type LeaderStamp struct {
	Index        int `json:"Index"`
	Term         int `json:"Term"`
	LeadersTicks int `json:"LeadersTicks"`
}
