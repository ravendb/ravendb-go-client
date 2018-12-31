package ravendb

import "sync"

// Lazy represents a lazy operation
type Lazy struct {
	valueFactory  func() (interface{}, error)
	valueFactory2 func(interface{}) error
	valueCreated  bool
	value         interface{}
	err           error
	mu            sync.Mutex
}

// NewLazy returns new Lazy value
func NewLazy(valueFactory func() (interface{}, error)) *Lazy {
	return &Lazy{
		valueFactory: valueFactory,
	}
}

// NewLazy2 returns new Lazy value
func NewLazy2(result interface{}, valueFactory func(interface{}) error) *Lazy {
	return &Lazy{
		value:         result,
		valueFactory2: valueFactory,
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

	if l.err != nil {
		panicIf(!l.valueCreated, "l.valuerCreated must have been set")
		return nil, l.err
	}

	if !l.valueCreated {
		l.value, l.err = l.valueFactory()
		l.valueCreated = true
	}
	return l.value, l.err
}

// GetValue2 executes lazy operation and ensures the value is set in result variable
// provided in NewLazy2()
func (l *Lazy) GetValue2() error {
	panicIf(l.valueFactory != nil, "for GetValue2 must create with NewLazy2")
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.valueCreated {
		return l.err
	}

	l.err = l.valueFactory2(l.value)
	l.valueCreated = true

	return l.err
}
