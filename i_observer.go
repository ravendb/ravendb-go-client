package ravendb

type IObserver interface {
	OnNext(interface{})
	OnError(error)
	OnCompleted()
}
