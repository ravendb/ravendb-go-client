package ravendb

// DeleteDatabaseResult represents result of Delete Database command
type DeleteDatabaseResult struct {
	RaftCommandIndex int64    `json:"RaftCommandIndex"`
	PendingDeletes   []string `json:"PendingDeletes"`
}
