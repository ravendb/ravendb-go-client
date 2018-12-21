package ravendb

// ConnectionString represents connection string
// TODO: unused
type ConnectionString struct {
	Name string
	// Note: Java has this as a virtual function getType()
	Type ConnectionStringType
}
