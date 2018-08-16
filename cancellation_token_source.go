package ravendb

import "time"

type CancellationTokenSource struct {
	cancelled bool

	cancelAfterDate time.Time // in milliseconds
}

func NewCancellationTokenSource() *CancellationTokenSource {
	return &CancellationTokenSource{}
}

func (s *CancellationTokenSource) getToken() *CancellationToken {
	return NewCancellationToken(s)
}

func NewCancellationToken(tokenSource *CancellationTokenSource) *CancellationToken {
	return &CancellationToken{
		token: tokenSource,
	}
}

func (s *CancellationTokenSource) cancel() {
	s.cancelled = true
}

func (s *CancellationTokenSource) cancelAfter(timeoutInMillis int) {
	dur := time.Millisecond * time.Duration(timeoutInMillis)
	s.cancelAfterDate = time.Now().Add(dur)
}

type CancellationToken struct {
	token *CancellationTokenSource
}

func (t *CancellationToken) isCancellationRequested() bool {
	if t.token.cancelled {
		return true
	}
	return !t.token.cancelAfterDate.IsZero() && time.Now().Sub(t.token.cancelAfterDate) > 0
}

func (t *CancellationToken) throwIfCancellationRequested() error {
	if t.isCancellationRequested() {
		return NewOperationCancelledException("")
	}
	return nil
}
