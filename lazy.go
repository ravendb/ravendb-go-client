package ravendb

import "sync"

type Lazy struct {
	valueFactory func() (interface{}, error)
	valueCreated bool
	value        interface{}
	mu           sync.Mutex
}

func NewLazy(valueFactory func() (interface{}, error)) *Lazy {
	return &Lazy{
		valueFactory: valueFactory,
	}
}

func (l *Lazy) IsValueCreated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.valueCreated
}

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
