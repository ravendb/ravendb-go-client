package ravendb

import "time"

const (
	// time format returned by the server
	// 2018-05-08T05:20:31.5233900Z
	serverTimeFormat = "2006-01-02T15:04:05.999999999Z"
)

// HiLoResult is a result of HiLoResult command
type HiLoResult struct {
	Prefix      string `json:"Prefix"`
	Low         int    `json:"Low"`
	High        int    `json:"High"`
	LastSize    int    `json:"LastSize"`
	ServerTag   string `json:"ServerTag"`
	LastRangeAt string `json:"LastRangeAt"`
}

// GetLastRangeAt parses LastRangeAt which is in a format:
// 2018-05-08T05:20:31.5233900Z
func (r *HiLoResult) GetLastRangeAt() time.Time {
	t, err := time.Parse(serverTimeFormat, r.LastRangeAt)
	must(err) // TODO: should silently fail? return an error?
	return t
}
