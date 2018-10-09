package ravendb

type ReplicationNode struct {
	URL        string `json:"Url"`
	Database   string `json:"Database"`
	IsDisabled bool   `json:"Disabled"`
}
