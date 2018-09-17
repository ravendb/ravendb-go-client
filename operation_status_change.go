package ravendb

// OperationStatusChange describes a change to the operation status. Can be used as DatabaseChange.
type OperationStatusChange struct {
	OperationID int
	State       map[string]interface{}
}
