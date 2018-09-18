package ravendb

type CleanCloseable interface {
	Close()
}

// NilCleanCloseable is meant for functions that return CleanCloseable type and
// want to return nil.
// Instead do:
// var res *NilCleanCloseable
// return res
// That way the caller can call Close() without checking for nil
type NilCleanCloseable struct {
}

func (n *NilCleanCloseable) Close() {
	// works even if n is nil
}

type FuncCleanCloseable struct {
	fn func()
}

func NewFuncCleanCloseable(fn func()) *FuncCleanCloseable {
	return &FuncCleanCloseable{
		fn: fn,
	}
}

func (f *FuncCleanCloseable) Close() {
	f.fn()
}
