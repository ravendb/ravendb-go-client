package ravendb

type IObserver interface {
	onNext(interface{})
	onError(error)
	onCompleted()
}
