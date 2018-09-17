package ravendb

var (
	// verify DatabaseConnectionState implements IChangesConnectionState
	_ IChangesConnectionState = &DatabaseConnectionState{}
)

type DatabaseConnectionState struct {
	onError []func(error)

	_onDisconnect Runnable
	onConnect     Runnable

	_value        atomicInteger
	lastException error

	onDocumentChangeNotification        []func(*DocumentChange)
	onIndexChangeNotification           []func(*IndexChange)
	onOperationStatusChangeNotification []func(*OperationStatusChange)
}

func (s *DatabaseConnectionState) addOnError(handler func(error)) int {
	s.onError = append(s.onError, handler)
	return len(s.onError) - 1
}

func (s *DatabaseConnectionState) removeOnError(idx int) {
	s.onError[idx] = nil
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
	for _, f := range s.onError {
		if f != nil {
			f(e)
		}
	}
}

func (s *DatabaseConnectionState) Close() {
	s.onDocumentChangeNotification = nil
	s.onIndexChangeNotification = nil
	s.onOperationStatusChangeNotification = nil
	s.onError = nil
}

func NewDatabaseConnectionState(onConnect Runnable, onDisconnect Runnable) *DatabaseConnectionState {
	return &DatabaseConnectionState{
		onConnect:     onConnect,
		_onDisconnect: onDisconnect,
	}
}

func (s *DatabaseConnectionState) addOnChangeNotification(typ ChangesType, handler interface{}) int {

	switch typ {
	case ChangesType_DOCUMENT:
		h := handler.(func(*DocumentChange))
		s.onDocumentChangeNotification = append(s.onDocumentChangeNotification, h)
		return len(s.onDocumentChangeNotification) - 1
	case ChangesType_INDEX:
		h := handler.(func(*IndexChange))
		s.onIndexChangeNotification = append(s.onIndexChangeNotification, h)
		return len(s.onIndexChangeNotification) - 1
	case ChangesType_OPERATION:
		h := handler.(func(*OperationStatusChange))
		s.onOperationStatusChangeNotification = append(s.onOperationStatusChangeNotification, h)
		return len(s.onOperationStatusChangeNotification) - 1
	default:
		//throw new IllegalStateException("ChangeType: " + type + " is not supported");
		panicIf(true, "ChangeType: %s is not supported", typ)
	}
	return -1
}

func (s *DatabaseConnectionState) removeOnChangeNotification(typ ChangesType, idx int) {
	switch typ {
	case ChangesType_DOCUMENT:
		s.onDocumentChangeNotification[idx] = nil

	case ChangesType_INDEX:
		s.onIndexChangeNotification[idx] = nil
	case ChangesType_OPERATION:
		s.onOperationStatusChangeNotification[idx] = nil
	default:
		//throw new IllegalStateException("ChangeType: " + type + " is not supported");
		panicIf(true, "ChangeType: %s is not supported", typ)
	}
}

func (s *DatabaseConnectionState) sendDocumentChange(documentChange *DocumentChange) {
	for _, f := range s.onDocumentChangeNotification {
		if f != nil {
			f(documentChange)
		}
	}
}

func (s *DatabaseConnectionState) sendIndexChange(indexChange *IndexChange) {
	for _, f := range s.onIndexChangeNotification {
		if f != nil {
			f(indexChange)
		}
	}
}

func (s *DatabaseConnectionState) sendOperationStatusChange(operationStatusChange *OperationStatusChange) {
	for _, f := range s.onOperationStatusChangeNotification {
		if f != nil {
			f(operationStatusChange)
		}
	}
}
