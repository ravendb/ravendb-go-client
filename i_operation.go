package ravendb

type IOperation interface {
	getCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand
}
