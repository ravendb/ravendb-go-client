package ravendb

var (
	// verify DatabaseConnectionState implements IChangesConnectionState
	_ IChangesConnectionState = &DatabaseConnectionState{}
)

type DatabaseConnectionState struct {
	onError []func(error)

	_onDisconnect Runnable
	onConnect     Runnable

	_value        AtomicInteger
	lastException error

	onDocumentChangeNotification []func(*DocumentChange)

	onIndexChangeNotification []func(*IndexChange)

	onOperationStatusChangeNotification []func(*OperationStatusChange)
}

func (s *DatabaseConnectionState) addOnError(handler func(error)) {
	s.onError = append(s.onError, handler)
}

func (s *DatabaseConnectionState) removeOnError(handler func(error)) {
	panicIf(true, "NYI")
	//s.onError = removeConsumerFromSlice(s.onError, handler)
}

func (s *DatabaseConnectionState) inc() {
	s._value.incrementAndGet()
}

func (s *DatabaseConnectionState) dec() {
	if s._value.decrementAndGet() == 0 {
		s._onDisconnect.run()
	}
}

func (s *DatabaseConnectionState) error(e error) {
	s.lastException = e
	panicIf(true, "NYI")
	//EventHelper.invoke(onError, e);
}

func (s *DatabaseConnectionState) Close() {
	panicIf(true, "NYI")

	//onDocumentChangeNotification.clear()
	//onIndexChangeNotification.clear()
	//onOperationStatusChangeNotification.clear()
	//onError.clear()
}

func NewDatabaseConnectionState(onConnect Runnable, onDisconnect Runnable) *DatabaseConnectionState {
	return &DatabaseConnectionState{
		onConnect:     onConnect,
		_onDisconnect: onDisconnect,
	}
}

func (s *DatabaseConnectionState) addOnChangeNotification(typ ChangesType, handler func(*DocumentChange)) {
	panicIf(true, "NYI")
}

func (s *DatabaseConnectionState) removeOnChangeNotification(typ ChangesType, handler func(*DocumentChange)) {
	panicIf(true, "NYI")
}

func (s *DatabaseConnectionState) sendDocumentChange(documentChange *DocumentChange) {
	for _, f := range s.onDocumentChangeNotification {
		f(documentChange)
	}
}

func (s *DatabaseConnectionState) sendIndexChange(indexChange *IndexChange) {
	for _, f := range s.onIndexChangeNotification {
		f(indexChange)
	}
}

func (s *DatabaseConnectionState) sendOperationStatusChange(operationStatusChange *OperationStatusChange) {
	for _, f := range s.onOperationStatusChangeNotification {
		f(operationStatusChange)
	}
}
