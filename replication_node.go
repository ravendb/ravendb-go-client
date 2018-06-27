package ravendb

type ReplicationNode struct {
	url      string `json:"Url"`
	database string `json:"Database"`
	disabled bool   `json:"Disabled"`
}

func (n *ReplicationNode) getUrl() string {
	return n.url
}

func (n *ReplicationNode) setUrl(url string) {
	n.url = url
}

func (n *ReplicationNode) getDatabase() string {
	return n.database
}

func (n *ReplicationNode) setDatabase(database string) {
	n.database = database
}

func (n *ReplicationNode) isDisabled() bool {
	return n.disabled
}

func (n *ReplicationNode) setDisabled(disabled bool) {
	n.disabled = disabled
}
