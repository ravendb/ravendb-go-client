package ravendb

type QueryTimings struct {
	DurationInMs int64                    `json:"DurationInMs"`
	Timings      map[string]*QueryTimings `json:"Timings"`
}

func (t *QueryTimings) update(queryResult *QueryResult) {
	t.DurationInMs = 0
	t.Timings = nil
	if queryResult.Timings == nil {
		return
	}
	t.DurationInMs = queryResult.Timings.DurationInMs
	t.Timings = queryResult.Timings.Timings
}
