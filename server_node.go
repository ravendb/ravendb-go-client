package ravendb

const (
	// ServerNode.Role
	ServerNode_Role_NONE       = "none"
	ServerNode_Role_PROMOTABLE = "promotable"
	ServerNode_Role_MEMBER     = "member"
	ServerNode_Role_REHAB      = "rehab"
)

// ServerNode describes a single server node
type ServerNode struct {
	URL        string `json:"Url"`
	Database   string `json:"Database"`
	ClusterTag string `json:"ClusterTag"`
	ServerRole string `json:"ServerRole"`
}

// NewServerNode creates a new ServerNode
func NewServerNode() *ServerNode {
	return &ServerNode{
		ServerRole: ServerNode_Role_NONE,
	}
}

func (n *ServerNode) getUrl() string {
	return n.URL
}

func (n *ServerNode) getDatabase() string {
	return n.Database
}

func (n *ServerNode) getServerRole() string {
	return n.ServerRole
}

func (n *ServerNode) setDatabase(database string) {
	n.Database = database
}

func (n *ServerNode) setUrl(url string) {
	n.URL = url
}
