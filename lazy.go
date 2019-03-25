package ravendb

import "sync"

// Lazy represents a lazy operation
type Lazy struct {
	// function which, when called, executes lazy operation
	valueFactory func(interface{}) error
	err          error
	valueCreated bool
	Value        interface{}
	mu           sync.Mutex
}

func newLazy(valueFactory func(interface{}) error) *Lazy {
	return &Lazy{
		valueFactory: valueFactory,
	}
}

// IsValueCreated returns true if lazy value has been created
func (l *Lazy) IsValueCreated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.err != nil {
		return false
	}

	return l.valueCreated
}

// GetValue executes lazy operation and ensures the Value is set in result variable
// provided in NewLazy()
func (l *Lazy) GetValue(result interface{}) error {
	if result == nil {
		return newIllegalArgumentError("result cannot be nil")
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.valueCreated {
		l.err = l.valueFactory(result)
		l.valueCreated = true
		if l.err != nil {
			l.Value = result
		}
	}

	if l.err != nil {
		return l.err
	}

	if l.Value == nil {
		return nil
	}

	// can call with nil to force evaluation of lazy operations
	if result != nil {
		setInterfaceToValue(result, l.Value)
	}

	return nil
}
