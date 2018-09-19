package ravendb

type IChangesObservable interface {
	Subscribe(observer IObserver) CleanCloseable
}
