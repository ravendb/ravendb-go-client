package ravendb

import "time"

// TODO: is time.Time here our *Time?
// TODO: needs json annotations?
type QueryStatistics struct {
	IsStale        bool
	DurationInMs   int64
	TotalResults   int
	SkippedResults int
	Timestamp      time.Time
	IndexName      string
	IndexTimestamp time.Time
	LastQueryTime  time.Time
	ResultEtag     int64 // TODO: *int64 ?
	ResultSize     int64
	NodeTag        string
}

func NewQueryStatistics() *QueryStatistics {
	return &QueryStatistics{}
}

func (s *QueryStatistics) UpdateQueryStats(qr *QueryResult) {
	s.IsStale = qr.IsStale
	s.DurationInMs = qr.DurationInMs
	s.TotalResults = qr.TotalResults
	s.SkippedResults = qr.SkippedResults
	s.Timestamp = qr.IndexTimestamp.toTime()
	s.IndexName = qr.IndexName
	s.IndexTimestamp = qr.IndexTimestamp.toTime()
	s.LastQueryTime = qr.LastQueryTime.toTime()
	s.ResultSize = qr.ResultSize
	s.ResultEtag = qr.ResultEtag
}
