package ravendb

import (
	"sync/atomic"
	"time"
)

// TODO: cancellationToken seems un-necessary

type cancellationTokenSource struct {
	cancelled int32

	timeDeadlineNanoSec int64
}

func (s *cancellationTokenSource) getToken() *cancellationToken {
	return &cancellationToken{
		token: s,
	}
}

func (s *cancellationTokenSource) cancel() {
	atomic.StoreInt32(&s.cancelled, 1)
}

/* TODO: remove
func (s *cancellationTokenSource) cancelAfter(timeoutInMilliseconds int) {
	dur := time.Millisecond * time.Duration(timeoutInMilliseconds)
	t := time.Now().Add(dur)
	atomic.StoreInt64(&s.timeDeadlineNanoSec, t.UnixNano())
}
*/

type cancellationToken struct {
	token *cancellationTokenSource
}

func (t *cancellationToken) isCancellationRequested() bool {
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

func (t *cancellationToken) throwIfCancellationRequested() error {
	if t.isCancellationRequested() {
		return newOperationCancelledError("")
	}
	return nil
}
