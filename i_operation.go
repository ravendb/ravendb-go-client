package ravendb

type IOperation interface {
	GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand
}
