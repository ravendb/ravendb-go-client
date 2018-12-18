package ravendb

// ModifyOngoingTaskResult represents a raven server command
// for modyfing task result
type ModifyOngoingTaskResult struct {
	TaskID           int    `json:"TaskId"`
	RaftCommandIndex int    `json:"RaftCommandIndex"`
	ResponsibleNode  string `json:"ResponsibleNode"`
}
