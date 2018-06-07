package ravendb

type IChangesConnectionState interface {
	// from CleanCloseable
	close()

	// IChangesConnectionState
	inc()
	dec()
	error(error)
	addOnChangeNotification(ChangesType, Consumer)
	removeOnChangeNotification(ChangesType, Consumer)
	addOnError(Consumer)
	removeOnError(Consumer)
}
