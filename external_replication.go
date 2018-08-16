package ravendb

type ExternalReplication struct {
	ReplicationNode
	TaskId               int    `json:"TaskId"`
	Name                 string `json:"Name"`
	ConnectionstringName string `json:"ConnectionstringName"`
	MentorName           string `json:"MentorName"`
}

func NewExternalReplication(database string, connectionstringName string) *ExternalReplication {
	r := &ExternalReplication{}
	r.setDatabase(database)
	r.setConnectionstringName(connectionstringName)
	return r
}

func (r *ExternalReplication) getTaskId() int {
	return r.TaskId
}

func (r *ExternalReplication) setTaskId(taskId int) {
	r.TaskId = taskId
}

func (r *ExternalReplication) GetName() string {
	return r.Name
}

func (r *ExternalReplication) setName(name string) {
	r.Name = name
}

func (r *ExternalReplication) getConnectionstringName() string {
	return r.ConnectionstringName
}

func (r *ExternalReplication) setConnectionstringName(connectionstringName string) {
	r.ConnectionstringName = connectionstringName
}

func (r *ExternalReplication) getMentorName() string {
	return r.MentorName
}

func (r *ExternalReplication) setMentorName(mentorName string) {
	r.MentorName = mentorName
}
