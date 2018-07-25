package ravendb

type IChangesConnectionState interface {
	// from CleanCloseable
	Close()

	// IChangesConnectionState
	inc()
	dec()
	error(error)
	addOnChangeNotification(ChangesType, func(*DocumentChange))
	removeOnChangeNotification(ChangesType, func(*DocumentChange))
	addOnError(func(error))
	removeOnError(func(error))
}
