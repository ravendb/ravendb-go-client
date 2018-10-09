package ravendb

type ExternalReplication struct {
	ReplicationNode
	TaskID               int    `json:"TaskId"`
	Name                 string `json:"Name"`
	ConnectionstringName string `json:"ConnectionstringName"`
	MentorName           string `json:"MentorName"`
}

func NewExternalReplication(database string, connectionstringName string) *ExternalReplication {
	return &ExternalReplication{
		ReplicationNode: ReplicationNode{
			Database: database,
		},
		ConnectionstringName: connectionstringName,
	}
}
