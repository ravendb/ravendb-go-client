package ravendb

// Note: we use io.Closer instead of CleanCloseable

// nilCloser is meant for functions that return io.Closer type and
// want to return nil.
// Instead do:
// var res *nilCloser
// return res
// That way the caller can call Close() without checking for nil
type nilCloser struct {
}

// Close closes nil closer
func (n *nilCloser) Close() error {
	// works even if n is nil
	return nil
}

// funcCloser wraps a function as io.Closer
type funcCloser struct {
	fn func() error
}

// newFuncCloser returns a new funcCloser
func newFuncCloser(fn func() error) *funcCloser {
	return &funcCloser{
		fn: fn,
	}
}

// Close calls underlying Close function
func (f *funcCloser) Close() error {
	return f.fn()
}
