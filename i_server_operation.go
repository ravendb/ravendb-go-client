package ravendb

type IServerOperation interface {
	getCommand(*DocumentConventions) RavenCommand
}
