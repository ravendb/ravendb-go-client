package ravendb

// ReplicationNode describes replication node
type ReplicationNode struct {
	URL        string `json:"Url"`
	Database   string `json:"Database"`
	IsDisabled bool   `json:"Disabled"`
}
