package ravendb

// OperationIDResult is a result of commands like CompactDatabaseCommand
type OperationIDResult struct {
	OperationID int64 `json:"OperationId"`
}
