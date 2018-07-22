package ravendb

import "time"

type Stopwatch struct {
	start time.Time
	dur   time.Duration
}

func Stopwatch_createStarted() *Stopwatch {
	return &Stopwatch{
		start: time.Now(),
	}
}

func (w *Stopwatch) stop() time.Duration {
	w.dur = time.Since(w.start)
	return w.dur
}

func (w *Stopwatch) String() string {
	return w.dur.String()
}
