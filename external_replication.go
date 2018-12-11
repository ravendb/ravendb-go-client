package ravendb

// ExternalReplication describes external replication
type ExternalReplication struct {
	ReplicationNode
	TaskID               int    `json:"TaskId"`
	Name                 string `json:"Name"`
	ConnectionstringName string `json:"ConnectionStringName"`
	MentorName           string `json:"MentorName"`
}

// NewExternalReplication creates ExternalReplication
func NewExternalReplication(database string, connectionstringName string) *ExternalReplication {
	return &ExternalReplication{
		ReplicationNode: ReplicationNode{
			Database: database,
		},
		ConnectionstringName: connectionstringName,
	}
}
