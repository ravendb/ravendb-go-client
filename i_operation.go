package ravendb

type IOperation interface {
	GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand
}
