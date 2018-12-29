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
func NewLazy2(valueFactory func(interface{}) error) *Lazy {
	return &Lazy{
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

// GetValue2 returns a value of lazy operation
func (l *Lazy) GetValue2(v interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.err != nil {
		panicIf(!l.valueCreated, "l.valuerCreated must have been set")
		return l.err
	}

	if l.valueCreated {
		panic("not yet supporting calling multiple times")

	}
	if !l.valueCreated {
		l.err = l.valueFactory2(v)
		l.value = v // TODO: this should convert **struct => *struct
		l.valueCreated = true
	} else {
		setInterfaceToValue(v, l.value)
	}

	return l.err
}
