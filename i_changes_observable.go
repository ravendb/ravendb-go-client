package ravendb

type IChangesObservable interface {
	subscribe(observer IObserver) CleanCloseable
}

type IChangesObservable_OperationStatusChange interface {
	subscribe(observer IObserver_OperationStatusChange) CleanCloseable
}
