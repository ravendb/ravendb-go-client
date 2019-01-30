package ravendb

// ConnectionString represents connection string
// used as argument to PutConnectionStringCommand
type ConnectionString struct {
	Name string `json:"Name"`
	// Note: Java has this as a virtual function getType()
	Type ConnectionStringType `json:"Type"`
}
