package ravendb

import "time"

// TODO: make private if not exposed in public API
// TODO: CancellationToken seems un-necessary

type CancellationTokenSource struct {
	cancelled bool

	cancelAfterDate time.Time
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
	s.cancelled = true
}

func (s *CancellationTokenSource) cancelAfter(timeoutInMilliseconds int) {
	dur := time.Millisecond * time.Duration(timeoutInMilliseconds)
	s.cancelAfterDate = time.Now().Add(dur)
}

type CancellationToken struct {
	token *CancellationTokenSource
}

func (t *CancellationToken) isCancellationRequested() bool {
	if t.token.cancelled {
		return true
	}
	if t.token.cancelAfterDate.IsZero() {
		return false
	}
	return time.Now().After(t.token.cancelAfterDate)
}

func (t *CancellationToken) throwIfCancellationRequested() error {
	if t.isCancellationRequested() {
		return NewOperationCancelledException("")
	}
	return nil
}
