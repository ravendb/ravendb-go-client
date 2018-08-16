package ravendb

// LeaderStamp describes leader stamp
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/serverwide/LeaderStamp.java#L3
type LeaderStamp struct {
	Index        int `json:"Index"`
	Term         int `json:"Term"`
	LeadersTicks int `json:"LeadersTicks"`
}
