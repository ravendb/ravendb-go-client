package ravendb

type IOperation interface {
	GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *httpCache) (RavenCommand, error)
}
