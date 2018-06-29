package ravendb

type ReplicationNode struct {
	Url      string `json:"Url"`
	Database string `json:"Database"`
	Disabled bool   `json:"Disabled"`
}

func (n *ReplicationNode) getUrl() string {
	return n.Url
}

func (n *ReplicationNode) setUrl(url string) {
	n.Url = url
}

func (n *ReplicationNode) getDatabase() string {
	return n.Database
}

func (n *ReplicationNode) setDatabase(database string) {
	n.Database = database
}

func (n *ReplicationNode) isDisabled() bool {
	return n.Disabled
}

func (n *ReplicationNode) setDisabled(disabled bool) {
	n.Disabled = disabled
}
