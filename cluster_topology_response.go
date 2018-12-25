package ravendb

// ClusterTopologyResponse is a response of GetClusterTopologyCommand
// Sample response:
// {"Topology":{"TopologyId":"8bf47de1-601e-4fff-b300-2e2c07ab6822","AllNodes":{"A":"http://localhost:9999"},"Members":{"A":"http://localhost:9999"},"Promotables":{},"Watchers":{},"LastNodeId":"A"},"Leader":"A","LeaderShipDuration":61407928,"CurrentState":"Leader","NodeTag":"A","CurrentTerm":4,"NodeLicenseDetails":{"A":{"UtilizedCores":3,"NumberOfCores":8,"InstalledMemoryInGb":16.0,"UsableMemoryInGb":16.0}},"LastStateChangeReason":"Leader, I'm the only one in thecluster, so I'm the leader (at 5/2/18 7:49:23 AM)","Status":{}}
type ClusterTopologyResponse struct {
	Leader   string           `json:"Leader"`
	NodeTag  string           `json:"NodeTag"`
	Topology *ClusterTopology `json:"Topology"`
	// note: the response returns more info
	// see https://app.quicktype.io?share=pzquGxXJcXyMncfA9JPa for fuller definition
}
