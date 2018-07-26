package ravendb

type IChangesConnectionState interface {
	// from CleanCloseable
	Close()

	// IChangesConnectionState
	inc()
	dec()
	error(error)
	addOnChangeNotification(ChangesType, interface{})
	removeOnChangeNotification(ChangesType, interface{})
	addOnError(func(error))
	removeOnError(func(error))
}
