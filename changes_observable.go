package ravendb

import "sync"

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

func (o *ChangesObservable) subscribe(observer IObserver) CleanCloseable {
	consumer := func(payload interface{}) {
		o.send(payload)
	}

	onErrorHandle := func(ex error) {
		o.error(ex)
	}

	consumerIdx := o._connectionState.addOnChangeNotification(o._type, consumer)
	onErrorHandleIdx := o._connectionState.addOnError(onErrorHandle)

	o._connectionState.inc()

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
	o.mu.Unlock()
}

func (o *ChangesObservable) removeObserver(observer IObserver) {
	o.mu.Lock()
	delete(o._subscribers, observer)
	o.mu.Unlock()
}

func (o *ChangesObservable) send(msg interface{}) {
	if o._filter != nil && o._filter(msg) {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	for subscriber := range o._subscribers {
		subscriber.onNext(msg)
	}
}

func (o *ChangesObservable) error(e error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for subscriber := range o._subscribers {
		subscriber.onError(e)
	}
}
