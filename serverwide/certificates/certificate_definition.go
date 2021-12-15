package certificates

type SecurityClearance int

const (
	ClusterAdmin SecurityClearance = iota
	ClusterNode
	Operator
	ValidUser
)

type DatabaseAccess int

const (
	ReadWrite DatabaseAccess = iota
	Admin
	Read
)

func (da DatabaseAccess) String() string {
	return []string{"ReadWrite", "Admin", "Read"}[da]
}

func (sa SecurityClearance) String() string {
	return []string{"ClusterAdmin", "ClusterNode", "Operator", "ValidUser"}[sa]
}
