package ravendb

// RavenConnectionString represents connection string for raven
type RavenConnectionString struct {
	ConnectionString
	Database              string   `json:"Database"`
	TopologyDiscoveryUrls []string `json:"TopologyDiscoveryUrls"`
}

func NewRavenConnectionString() *RavenConnectionString {
	res := &RavenConnectionString{}
	res.Type = ConnectionStringTypeRaven
	return res
}
