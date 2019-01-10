package ravendb

import "time"

type stopWatch struct {
	start time.Time
	dur   time.Duration
}

func newStopWatchStarted() *stopWatch {
	return &stopWatch{
		start: time.Now(),
	}
}

func (w *stopWatch) stop() time.Duration {
	w.dur = time.Since(w.start)
	return w.dur
}

func (w *stopWatch) String() string {
	return w.dur.String()
}
