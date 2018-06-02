package ravendb

// HiLoResult is a result of HiLoResult command
type HiLoResult struct {
	Prefix      string      `json:"Prefix"`
	Low         int         `json:"Low"`
	High        int         `json:"High"`
	LastSize    int         `json:"LastSize"`
	ServerTag   string      `json:"ServerTag"`
	LastRangeAt *ServerTime `json:"LastRangeAt"`
}
