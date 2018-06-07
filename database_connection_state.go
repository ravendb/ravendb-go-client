package ravendb

var (
	// verify DatabaseConnectionState implements IChangesConnectionState
	_ IChangesConnectionState = &DatabaseConnectionState{}
)

type DatabaseConnectionState struct {
	onError []Consumer

	_onDisconnect Runnable
	onConnect     Runnable

	_value        AtomicInteger
	lastException error

	onDocumentChangeNotification []Consumer

	onIndexChangeNotification []Consumer

	onOperationStatusChangeNotification []Consumer
}

func (s *DatabaseConnectionState) addOnError(handler Consumer) {
	s.onError = append(s.onError, handler)
}

func (s *DatabaseConnectionState) removeOnError(handler Consumer) {
	s.onError = removeConsumerFromSlice(s.onError, handler)
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

func (s *DatabaseConnectionState) close() {
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

func (s *DatabaseConnectionState) addOnChangeNotification(typ ChangesType, handler Consumer) {
	panicIf(true, "NYI")
}

func (s *DatabaseConnectionState) removeOnChangeNotification(typ ChangesType, handler Consumer) {
	panicIf(true, "NYI")
}

func (s *DatabaseConnectionState) sendDocumentChange(documentChange *DocumentChange) {
	EventHelper_invoke(s.onDocumentChangeNotification, documentChange)
}

func (s *DatabaseConnectionState) sendIndexChange(indexChange *IndexChange) {
	EventHelper_invoke(s.onIndexChangeNotification, indexChange)
}

func (s *DatabaseConnectionState) sendOperationStatusChange(operationStatusChange *OperationStatusChange) {
	EventHelper_invoke(s.onOperationStatusChangeNotification, operationStatusChange)
}

func removeConsumerFromSlice(a []Consumer, toRemove Consumer) []Consumer {
	// TODO: optimize
	var res []Consumer
	for _, c := range a {
		if c != toRemove {
			res = append(res, c)
		}
	}
	return res
}
