package ravendb

type IMaintenanceOperation interface {
	GetCommand(*DocumentConventions) RavenCommand
}
