package ravendb

import (
	"fmt"
	"sync"
)

var (
	_ IChangesConnectionState = &DatabaseConnectionState{}
)

type DatabaseConnectionState struct {
	onError []func(error)

	_onDisconnect Runnable
	onConnect     Runnable

	_value        atomicInteger
	lastException error

	onDocumentChangeNotification        []func(interface{})
	onIndexChangeNotification           []func(interface{})
	onOperationStatusChangeNotification []func(interface{})

	// protects arrays
	mu sync.Mutex
}

func (s *DatabaseConnectionState) addOnError(handler func(error)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onError = append(s.onError, handler)
	return len(s.onError) - 1
}

func (s *DatabaseConnectionState) removeOnError(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onError[idx] = nil
}

func (s *DatabaseConnectionState) inc() {
	s._value.incrementAndGet()
}

func (s *DatabaseConnectionState) dec() {
	if s._value.decrementAndGet() == 0 {
		if s._onDisconnect != nil {
			s._onDisconnect()
		}
	}
}

func (s *DatabaseConnectionState) error(e error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastException = e
	for _, f := range s.onError {
		if f != nil {
			f(e)
		}
	}
}

func (s *DatabaseConnectionState) Close() {
	// Note: not clearing as in Java because removeOnChangeNotification()
	// can be called after Close()
}

func NewDatabaseConnectionState(onConnect Runnable, onDisconnect Runnable) *DatabaseConnectionState {
	return &DatabaseConnectionState{
		onConnect:     onConnect,
		_onDisconnect: onDisconnect,
	}
}

func (s *DatabaseConnectionState) addOnChangeNotification(typ ChangesType, handler func(interface{})) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	var idx int
	switch typ {
	case ChangesType_DOCUMENT:
		idx = len(s.onDocumentChangeNotification)
		s.onDocumentChangeNotification = append(s.onDocumentChangeNotification, handler)
	case ChangesType_INDEX:
		idx = len(s.onIndexChangeNotification)
		s.onIndexChangeNotification = append(s.onIndexChangeNotification, handler)
	case ChangesType_OPERATION:
		idx = len(s.onOperationStatusChangeNotification)
		s.onOperationStatusChangeNotification = append(s.onOperationStatusChangeNotification, handler)
	default:
		//throw new IllegalStateError("ChangeType: " + type + " is not supported");
		panicIf(true, "ChangeType: %s is not supported", typ)
	}
	return idx
}

func (s *DatabaseConnectionState) removeOnChangeNotification(typ ChangesType, idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch typ {
	case ChangesType_DOCUMENT:
		s.onDocumentChangeNotification[idx] = nil
	case ChangesType_INDEX:
		s.onIndexChangeNotification[idx] = nil
	case ChangesType_OPERATION:
		s.onOperationStatusChangeNotification[idx] = nil
	default:
		//throw new IllegalStateError("ChangeType: " + type + " is not supported");
		panicIf(true, "ChangeType: %s is not supported", typ)
	}
}

func (s *DatabaseConnectionState) send(v interface{}) error {
	switch rv := v.(type) {
	case *DocumentChange:
		s.sendDocumentChange(rv)
	case *IndexChange:
		s.sendIndexChange(rv)
	case *OperationStatusChange:
		s.sendOperationStatusChange(rv)
	default:
		return fmt.Errorf("DatabaseConnectionState.send(): unsupporrted type %T", v)
	}
	return nil
}

func (s *DatabaseConnectionState) sendDocumentChange(documentChange *DocumentChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onDocumentChangeNotification {
		if f != nil {
			f(documentChange)
		}
	}
}

func (s *DatabaseConnectionState) sendIndexChange(indexChange *IndexChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onIndexChangeNotification {
		if f != nil {
			f(indexChange)
		}
	}
}

func (s *DatabaseConnectionState) sendOperationStatusChange(operationStatusChange *OperationStatusChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onOperationStatusChangeNotification {
		if f != nil {
			f(operationStatusChange)
		}
	}
}
