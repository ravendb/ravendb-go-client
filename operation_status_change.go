package ravendb

// OperationStatusChange describes a change to the operation status. Can be used as DatabaseChange.
type OperationStatusChange struct {
	OperationID int64
	State       map[string]interface{}
}
