package ravendb

import "io"

type IObservable interface {
	subscribe(observer *IObserver) io.Closer
}
