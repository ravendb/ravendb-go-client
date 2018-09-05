package ravendb

// For documentation/porting only. Go has no generics so it's the same
// as IMaintenanceOperation
type IVoidMaintenanceOperation = IMaintenanceOperation

type IMaintenanceOperation interface {
	GetCommand(*DocumentConventions) RavenCommand
}
