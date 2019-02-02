package ravendb

type IServerOperation interface {
	GetCommand(*DocumentConventions) (RavenCommand, error)
}
