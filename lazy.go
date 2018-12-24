package ravendb

import "sync"

// Lazy represents a lazy operation
type Lazy struct {
	valueFactory func() (interface{}, error)
	valueCreated bool
	value        interface{}
	mu           sync.Mutex
}

// NewLazy returns new Lazy value
func NewLazy(valueFactory func() (interface{}, error)) *Lazy {
	return &Lazy{
		valueFactory: valueFactory,
	}
}

// IsValueCreated returns true if lazy value has been created
func (l *Lazy) IsValueCreated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.valueCreated
}

// GetValue returns a value of lazy operation
func (l *Lazy) GetValue() (interface{}, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var err error
	if !l.valueCreated {
		l.value, err = l.valueFactory()
		l.valueCreated = true
	}
	return l.value, err
}
