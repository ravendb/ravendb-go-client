package ravendb

// HiLoResult is a result of HiLoResult command
type HiLoResult struct {
	Prefix      string `json:"Prefix"`
	Low         int64  `json:"Low"`
	High        int64  `json:"High"`
	LastSize    int64  `json:"LastSize"`
	ServerTag   string `json:"ServerTag"`
	LastRangeAt *Time  `json:"LastRangeAt"`
}
