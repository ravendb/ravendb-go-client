package ravendb

import (
	"sync"
)

// Note: in Java IChangesConnectionState hides databaseConnectionState

type databaseConnectionState struct {
	onError []func(error)

	onDisconnect func()
	onConnect    func()

	value     atomicInteger
	lastError error

	onDocumentChangeNotification        []func(*DocumentChange)
	onIndexChangeNotification           []func(*IndexChange)
	onOperationStatusChangeNotification []func(*OperationStatusChange)

	// protects arrays
	mu sync.Mutex
}

func (s *databaseConnectionState) addOnError(handler func(error)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onError = append(s.onError, handler)
	return len(s.onError) - 1
}

func (s *databaseConnectionState) removeOnError(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onError[idx] = nil
}

func (s *databaseConnectionState) inc() {
	s.value.incrementAndGet()
}

func (s *databaseConnectionState) dec() {
	if s.value.decrementAndGet() == 0 {
		if s.onDisconnect != nil {
			s.onDisconnect()
		}
	}
}

func (s *databaseConnectionState) error(e error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = e
	for _, f := range s.onError {
		if f != nil {
			f(e)
		}
	}
}

// Close closes the connection
func (s *databaseConnectionState) Close() error {
	// Note: not clearing as in Java because removeOnChangeNotification()
	// can be called after Close()
	return nil
}

func newDatabaseConnectionState(onConnect func(), onDisconnect func()) *databaseConnectionState {
	return &databaseConnectionState{
		onConnect:    onConnect,
		onDisconnect: onDisconnect,
	}
}

func (s *databaseConnectionState) addOnDocumentChangeNotification(handler func(*DocumentChange)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onDocumentChangeNotification = append(s.onDocumentChangeNotification, handler)
	return len(s.onDocumentChangeNotification) - 1
}

func (s *databaseConnectionState) addOnIndexChangeNotification(handler func(*IndexChange)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onIndexChangeNotification = append(s.onIndexChangeNotification, handler)
	return len(s.onIndexChangeNotification) - 1
}

func (s *databaseConnectionState) addOnOperationChangeNotification(handler func(*OperationStatusChange)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onOperationStatusChangeNotification = append(s.onOperationStatusChangeNotification, handler)
	return len(s.onOperationStatusChangeNotification) - 1
}

func (s *databaseConnectionState) removeOnDocumentChangeNotification(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onDocumentChangeNotification[idx] = nil
}

func (s *databaseConnectionState) removeOnIndexChangeNotification(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onIndexChangeNotification[idx] = nil
}

func (s *databaseConnectionState) removeOnOperationChangeNotification(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onOperationStatusChangeNotification[idx] = nil
}

func (s *databaseConnectionState) sendDocumentChange(documentChange *DocumentChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onDocumentChangeNotification {
		if f != nil {
			f(documentChange)
		}
	}
}

func (s *databaseConnectionState) sendIndexChange(indexChange *IndexChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onIndexChangeNotification {
		if f != nil {
			f(indexChange)
		}
	}
}

func (s *databaseConnectionState) sendOperationStatusChange(operationStatusChange *OperationStatusChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onOperationStatusChangeNotification {
		if f != nil {
			f(operationStatusChange)
		}
	}
}
