package ravendb

// BatchCommandResult is a result of batch server command
type BatchCommandResult struct {
	Results          []map[string]interface{} `json:"Results"`
	TransactionIndex int64                    `json:"TransactionIndex"`
}
