package ravendb

import (
	"fmt"
	"sync"
)

var _ IChangesObservable = &ChangesObservable{}

type ChangesObservable struct {
	_type            ChangesType
	_connectionState IChangesConnectionState
	_filter          func(interface{}) bool
	mu               sync.Mutex
	_subscribers     map[IObserver]bool
}

func NewChangesObservable(typ ChangesType, connectionState IChangesConnectionState, filter func(interface{}) bool) *ChangesObservable {
	return &ChangesObservable{
		_type:            typ,
		_connectionState: connectionState,
		_filter:          filter,
		_subscribers:     make(map[IObserver]bool),
	}
}

func (o *ChangesObservable) Subscribe(observer IObserver) CleanCloseable {
	dbg("ChangesObservable.Subscribe\n")
	consumer := func(payload interface{}) {
		dbg("ChangesObservable.Subscribe: consumer called\n")
		o.send(payload)
	}

	onErrorHandle := func(ex error) {
		dbg("ChangesObservable.Subscribe: onErrorHandle called\n")
		o.error(ex)
	}

	consumerIdx := o._connectionState.addOnChangeNotification(o._type, consumer)
	onErrorHandleIdx := o._connectionState.addOnError(onErrorHandle)

	o._connectionState.inc()
	o.addObserver(observer)

	fn := func() {
		o._connectionState.dec()
		o.removeObserver(observer)
		o._connectionState.removeOnChangeNotification(o._type, consumerIdx)
		o._connectionState.removeOnError(onErrorHandleIdx)
	}
	return NewFuncCleanCloseable(fn)
}

func (o *ChangesObservable) addObserver(observer IObserver) {
	o.mu.Lock()
	o._subscribers[observer] = true
	dbg("ChangesObservable.send() addObserver() len(o._subscribers)=%d\n", len(o._subscribers))
	o.mu.Unlock()
}

func (o *ChangesObservable) removeObserver(observer IObserver) {
	o.mu.Lock()
	delete(o._subscribers, observer)
	dbg("ChangesObservable.send() removeObserver() len(o._subscribers)=%d\n", len(o._subscribers))
	o.mu.Unlock()
}

func (o *ChangesObservable) send(msg interface{}) {
	if o._filter != nil && !o._filter(msg) {
		fmt.Printf("ChangesObservable.send() skipping because filter returned false\n")
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	dbg("ChangesObservable.send() len(o._subscribers)=%d\n", len(o._subscribers))
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
