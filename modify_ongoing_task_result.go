package ravendb

// ModifyOngoingTaskResult represents a raven server command
// for modyfing task result
type ModifyOngoingTaskResult struct {
	TaskID           int64  `json:"TaskId"`
	RaftCommandIndex int64  `json:"RaftCommandIndex"`
	ResponsibleNode  string `json:"ResponsibleNode"`
}
