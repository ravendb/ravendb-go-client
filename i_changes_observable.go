package ravendb

import "io"

type IChangesObservable interface {
	Subscribe(observer IObserver) io.Closer
}
