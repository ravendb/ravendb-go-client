package ravendb

type ModifyOngoingTaskResult struct {
	TaskId           int    `json:"TaskId"`
	RaftCommandIndex int    `json:"RaftCommandIndex"`
	ResponsibleNode  string `json:"ResponsibleNode"`
}

func (r *ModifyOngoingTaskResult) getTaskId() int {
	return r.TaskId
}

func (r *ModifyOngoingTaskResult) setTaskId(taskId int) {
	r.TaskId = taskId
}

func (r *ModifyOngoingTaskResult) getRaftCommandIndex() int {
	return r.RaftCommandIndex
}

func (r *ModifyOngoingTaskResult) setRaftCommandIndex(raftCommandIndex int) {
	r.RaftCommandIndex = raftCommandIndex
}

func (r *ModifyOngoingTaskResult) getResponsibleNode() string {
	return r.ResponsibleNode
}

func (r *ModifyOngoingTaskResult) setResponsibleNode(responsibleNode string) {
	r.ResponsibleNode = responsibleNode
}
