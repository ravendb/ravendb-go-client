package ravendb

import (
	"fmt"
	"time"
)

// TODO: make private if not exposed in public API
// TODO: CancellationToken seems un-necessary
// TODO: should use atomic becasue used across threads

type CancellationTokenSource struct {
	cancelled bool

	cancelAfterDate time.Time
}

func NewCancellationTokenSource() *CancellationTokenSource {
	return &CancellationTokenSource{}
}

func (s *CancellationTokenSource) getToken() *CancellationToken {
	dbg("CancellationTokenSource.getToken()\n")
	return &CancellationToken{
		token: s,
	}
}

func (s *CancellationTokenSource) cancel() {
	fmt.Printf("token requested cancelled\n")
	s.cancelled = true
}

func (s *CancellationTokenSource) cancelAfter(timeoutInMilliseconds int) {
	dbg("token requested cancelled for %d ms\n", timeoutInMilliseconds)
	dur := time.Millisecond * time.Duration(timeoutInMilliseconds)
	s.cancelAfterDate = time.Now().Add(dur)
}

type CancellationToken struct {
	token *CancellationTokenSource
}

func (t *CancellationToken) isCancellationRequested() bool {
	if t.token.cancelled {
		dbg("CancellationToken.isCancellationRequested: yes because cancelled=%v\n", t.token.cancelled)
		return true
	}
	if t.token.cancelAfterDate.IsZero() {
		dbg("CancellationToken.isCancellationRequested: no because token.cancelAfterDate is zero\n")
		return false
	}
	isAfter := time.Now().After(t.token.cancelAfterDate)
	dbg("CancellationToken.isCancellationRequested: isAfter=%v\n", isAfter)
	return isAfter
}

func (t *CancellationToken) throwIfCancellationRequested() error {
	if t.isCancellationRequested() {
		return NewOperationCancelledException("")
	}
	return nil
}
