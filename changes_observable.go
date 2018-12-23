package ravendb

import (
	"io"
	"sync"
)

var _ IChangesObservable = &ChangesObservable{}

// ChangesObservable is for observing changes in a database
type ChangesObservable struct {
	_type            ChangesType
	_connectionState *DatabaseConnectionState
	_filter          func(interface{}) bool
	mu               sync.Mutex
	_subscribers     map[IObserver]bool
}

// NewChangesObservable returns a new ChangesObservable
func NewChangesObservable(typ ChangesType, connectionState *DatabaseConnectionState, filter func(interface{}) bool) *ChangesObservable {
	return &ChangesObservable{
		_type:            typ,
		_connectionState: connectionState,
		_filter:          filter,
		_subscribers:     make(map[IObserver]bool),
	}
}

// Subscribe subscribes to a given observer
func (o *ChangesObservable) Subscribe(observer IObserver) io.Closer {
	consumer := func(payload interface{}) {
		o.send(payload)
	}

	onErrorHandle := func(ex error) {
		o.error(ex)
	}

	consumerIdx := o._connectionState.addOnChangeNotification(o._type, consumer)
	onErrorHandleIdx := o._connectionState.addOnError(onErrorHandle)

	o._connectionState.inc()
	o.addObserver(observer)

	fn := func() error {
		o._connectionState.dec()
		o.removeObserver(observer)
		o._connectionState.removeOnChangeNotification(o._type, consumerIdx)
		o._connectionState.removeOnError(onErrorHandleIdx)
		return nil
	}
	return newFuncCloser(fn)
}

func (o *ChangesObservable) addObserver(observer IObserver) {
	o.mu.Lock()
	o._subscribers[observer] = true
	o.mu.Unlock()
}

func (o *ChangesObservable) removeObserver(observer IObserver) {
	o.mu.Lock()
	delete(o._subscribers, observer)
	o.mu.Unlock()
}

func (o *ChangesObservable) send(msg interface{}) {
	if o._filter != nil && !o._filter(msg) {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	for subscriber := range o._subscribers {
		subscriber.OnNext(msg)
	}
}

func (o *ChangesObservable) error(e error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for subscriber := range o._subscribers {
		subscriber.OnError(e)
	}
}
