package certificates

var SecurityClearance = newSecurityClearance()

func newSecurityClearance() *securityClearance {
	return &securityClearance{
		ClusterAdmin: "ClusterAdmin",
		ClusterNode:  "ClusterNode",
		Operator:     "Operator",
		ValidUser:    "ValidUser",
	}
}

type securityClearance = struct {
	ClusterAdmin string `default:"ClusterAdmin"`
	ClusterNode  string `default:"ClusterNode"`
	Operator     string `default:"Operator"`
	ValidUser    string `default:"ValidUser"`
}

var DatabaseAccess = newDatabaseAccess()

func newDatabaseAccess() *databaseAccess {
	return &databaseAccess{
		ReadWrite: "ReadWrite",
		Admin:     "Admin",
		Read:      "Read",
	}
}

type databaseAccess = struct {
	ReadWrite string `default:"ReadWrite"`
	Admin     string `default:"Admin"`
	Read      string `default:"Read"`
}
