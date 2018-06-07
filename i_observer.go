package ravendb

type IObserver interface {
	onNext(interface{})
	onError(error)
	onCompleted()
}

type IObserver_OperationStatusChange interface {
	onNext(*OperationStatusChange)
	onError(error)
	onCompleted()
}
