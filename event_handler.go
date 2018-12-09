package ravendb

type EventHandler interface {
	handle(sender interface{}, event interface{})
}
