package ravendb

import "time"

type StreamQueryStatistics struct {
	IndexName      string
	IsStale        bool
	IndexTimestamp time.Time
	TotalResults   int
	ResultEtag     int64
}
