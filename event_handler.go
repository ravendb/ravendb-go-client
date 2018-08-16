package ravendb

type EventHandler interface {
	handle(sender Object, event interface{})
}
