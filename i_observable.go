package ravendb

type IObservable interface {
	subscribe(observer *IObserver) CleanCloseable
}
