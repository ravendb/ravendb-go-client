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

func (s *DatabaseConnectionState) addOnChangeNotification(typ ChangesType, handler interface{}) {

	switch typ {
	case ChangesType_DOCUMENT:
		h := handler.(func(*DocumentChange))
		s.onDocumentChangeNotification = append(s.onDocumentChangeNotification, h)
	case ChangesType_INDEX:
		h := handler.(func(*IndexChange))
		s.onIndexChangeNotification = append(s.onIndexChangeNotification, h)
	case ChangesType_OPERATION:
		h := handler.(func(*OperationStatusChange))
		s.onOperationStatusChangeNotification = append(s.onOperationStatusChangeNotification, h)
	default:
		//throw new IllegalStateException("ChangeType: " + type + " is not supported");
		panicIf(true, "ChangeType: %s is not supported", typ)
	}
}

func (s *DatabaseConnectionState) removeOnChangeNotification(typ ChangesType, handler interface{}) {
	switch typ {
	case ChangesType_DOCUMENT:
		curr := 0
		a := s.onDocumentChangeNotification
		n := len(a)
		for i, el := range a {
			var v interface{} = el
			if handler == v {
				continue
			}
			if i != curr {
				a[curr] = a[i]
			}
			curr++
		}
		if curr < n-1 {
			s.onDocumentChangeNotification = a[:curr]
		}

	case ChangesType_INDEX:
		curr := 0
		a := s.onIndexChangeNotification
		n := len(a)
		for i, el := range a {
			var v interface{} = el
			if handler == v {
				continue
			}
			if i != curr {
				a[curr] = a[i]
			}
			curr++
		}
		if curr < n-1 {
			s.onIndexChangeNotification = a[:curr]
		}
	case ChangesType_OPERATION:
		curr := 0
		a := s.onOperationStatusChangeNotification
		n := len(a)
		for i, el := range a {
			var v interface{} = el
			if handler == v {
				continue
			}
			if i != curr {
				a[curr] = a[i]
			}
			curr++
		}
		if curr < n-1 {
			s.onOperationStatusChangeNotification = a[:curr]
		}
	default:
		//throw new IllegalStateException("ChangeType: " + type + " is not supported");
		panicIf(true, "ChangeType: %s is not supported", typ)
	}
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
