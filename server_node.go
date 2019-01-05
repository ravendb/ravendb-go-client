package ravendb

const (
	ServerNodeRoleNone       = "None"
	ServerNodeRolePromotable = "Promotable"
	ServerNodeRoleMember     = "Member"
	ServerNodeRoleRehab      = "Rehab"
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
		ServerRole: ServerNodeRoleNone,
	}
}
