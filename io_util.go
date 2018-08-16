package ravendb

import (
	"bytes"
	"io"
)

// CapturingReadCloser is a reader that captures data that was read from
// underlying reader
type CapturingReadCloser struct {
	tee          io.Reader
	orig         io.ReadCloser
	capturedData bytes.Buffer
	wasClosed    bool
}

// Read reads data from reader
func (rc *CapturingReadCloser) Read(p []byte) (int, error) {
	panicIf(rc.wasClosed, "reading after being closed")
	return rc.tee.Read(p)
}

// Close closes a reader
func (rc *CapturingReadCloser) Close() error {
	rc.wasClosed = true
	return rc.orig.Close()
}

// NewCapturingReadCloser returns a new capturing reader
func NewCapturingReadCloser(orig io.ReadCloser) *CapturingReadCloser {
	res := &CapturingReadCloser{
		orig: orig,
	}
	res.tee = io.TeeReader(orig, &res.capturedData)
	return res
}
