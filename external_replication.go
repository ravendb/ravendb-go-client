package ravendb

type ExternalReplication struct {
	ReplicationNode
	taskId               int    `json:"TaskId"`
	name                 string `json:"Name"`
	connectionstringName string `json:"ConnectionstringName"`
	mentorName           string `json:"MentorName"`
}

func NewExternalReplication(database string, connectionstringName string) *ExternalReplication {
	r := &ExternalReplication{}
	r.setDatabase(database)
	r.setConnectionstringName(connectionstringName)
	return r
}

func (r *ExternalReplication) getTaskId() int {
	return r.taskId
}

func (r *ExternalReplication) setTaskId(taskId int) {
	r.taskId = taskId
}

func (r *ExternalReplication) getName() string {
	return r.name
}

func (r *ExternalReplication) setName(name string) {
	r.name = name
}

func (r *ExternalReplication) getConnectionstringName() string {
	return r.connectionstringName
}

func (r *ExternalReplication) setConnectionstringName(connectionstringName string) {
	r.connectionstringName = connectionstringName
}

func (r *ExternalReplication) getMentorName() string {
	return r.mentorName
}

func (r *ExternalReplication) setMentorName(mentorName string) {
	r.mentorName = mentorName
}
