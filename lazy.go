package ravendb

import "sync"

// Lazy represents a lazy operation
type Lazy struct {
	// function which, when called, executes lazy operation
	valueFactory func(interface{}) error
	valueCreated bool
	// passed by the user, where the result of lazy operation is stored
	// usually doesn't need to be read because the caller passed the result
	// in when calling NewLazy. But if you must read it, make sure to call
	// GetValue() first
	Value interface{}
	err   error
	mu    sync.Mutex
}

// NewLazy2 returns new Lazy value
func NewLazy(result interface{}, valueFactory func(interface{}) error) *Lazy {
	return &Lazy{
		Value:        result,
		valueFactory: valueFactory,
	}
}

// IsValueCreated returns true if lazy value has been created
func (l *Lazy) IsValueCreated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.valueCreated
}

// GetValue executes lazy operation and ensures the Value is set in result variable
// provided in NewLazy()
func (l *Lazy) GetValue() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.valueCreated {
		return l.err
	}

	l.err = l.valueFactory(l.Value)
	l.valueCreated = true

	return l.err
}
