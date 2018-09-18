package ravendb

type IChangesConnectionState interface {
	// from CleanCloseable
	Close()

	// IChangesConnectionState
	inc()
	dec()
	error(error)
	addOnChangeNotification(ChangesType, func(interface{})) int
	removeOnChangeNotification(ChangesType, int)
	addOnError(func(error)) int
	removeOnError(int)
}
