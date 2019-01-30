package ravendb

// ExternalReplication describes external replication
type ExternalReplication struct {
	ReplicationNode
	TaskID               int64  `json:"TaskId"`
	Name                 string `json:"Name"`
	ConnectionStringName string `json:"ConnectionStringName"`
	MentorName           string `json:"MentorName"`
}

// NewExternalReplication creates ExternalReplication
func NewExternalReplication(database string, connectionStringName string) *ExternalReplication {
	return &ExternalReplication{
		ReplicationNode: ReplicationNode{
			Database: database,
		},
		ConnectionStringName: connectionStringName,
	}
}
