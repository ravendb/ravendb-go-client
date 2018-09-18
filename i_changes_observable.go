package ravendb

type IChangesObservable interface {
	subscribe(observer IObserver) CleanCloseable
}
