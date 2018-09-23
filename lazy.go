package ravendb

import "sync"

type Lazy struct {
	valueFactory func() interface{}
	valueCreated bool
	value        interface{}
	mu           sync.Mutex
}

func NewLazy(valueFactory func() interface{}) *Lazy {
	return &Lazy{
		valueFactory: valueFactory,
	}
}

func (l *Lazy) isValueCreated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.valueCreated
}

func (l *Lazy) getValue() interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.valueCreated {
		l.value = l.valueFactory()
		l.valueCreated = true
	}
	return l.value
}
