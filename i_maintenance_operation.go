package ravendb

type IMaintenanceOperation interface {
	getCommand(*DocumentConventions) *RavenCommand
}
