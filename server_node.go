package ravendb

const (
	// ServerNode.Role
	ServerNode_Role_NONE       = "None"
	ServerNode_Role_PROMOTABLE = "Promotable"
	ServerNode_Role_MEMBER     = "Member"
	ServerNode_Role_REHAB      = "Rehab"
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

func (n *ServerNode) getClusterTag() string {
	return n.ClusterTag
}

func (n *ServerNode) setDatabase(database string) {
	n.Database = database
}

func (n *ServerNode) setUrl(url string) {
	n.URL = url
}

func (n *ServerNode) setClusterTag(clusterTag string) {
	n.ClusterTag = clusterTag
}
