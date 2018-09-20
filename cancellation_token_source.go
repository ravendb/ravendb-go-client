package ravendb

import (
	"sync/atomic"
	"time"
)

// TODO: make private if not exposed in public API
// TODO: CancellationToken seems un-necessary

type CancellationTokenSource struct {
	cancelled int32

	timeDeadlineNanoSec int64
}

func NewCancellationTokenSource() *CancellationTokenSource {
	return &CancellationTokenSource{}
}

func (s *CancellationTokenSource) getToken() *CancellationToken {
	return &CancellationToken{
		token: s,
	}
}

func (s *CancellationTokenSource) cancel() {
	atomic.StoreInt32(&s.cancelled, 1)
}

func (s *CancellationTokenSource) cancelAfter(timeoutInMilliseconds int) {
	dur := time.Millisecond * time.Duration(timeoutInMilliseconds)
	t := time.Now().Add(dur)
	atomic.StoreInt64(&s.timeDeadlineNanoSec, t.UnixNano())
}

type CancellationToken struct {
	token *CancellationTokenSource
}

func (t *CancellationToken) isCancellationRequested() bool {
	v := atomic.LoadInt32(&t.token.cancelled)
	if v != 0 {
		return true
	}
	timeDeadlineNanoSec := atomic.LoadInt64(&t.token.timeDeadlineNanoSec)
	if 0 == timeDeadlineNanoSec {
		return false
	}

	timeDeadline := time.Unix(0, timeDeadlineNanoSec)
	isAfter := time.Now().After(timeDeadline)
	return isAfter
}

func (t *CancellationToken) throwIfCancellationRequested() error {
	if t.isCancellationRequested() {
		return NewOperationCancelledException("")
	}
	return nil
}
