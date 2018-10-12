package ravendb

import "time"

type StreamQueryStatistics struct {
	IndexName      string    `json:"IndexName"`
	IsStale        bool      `json:"IsStale"`
	IndexTimestamp time.Time `json:"IndexTimestamp"`
	TotalResults   int       `json:"TotalResults"`
	ResultEtag     int64     `json:"ResultEtag"`
}
