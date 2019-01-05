package ravendb

// ClusterTopologyResponse is a response of GetClusterTopologyCommand
type ClusterTopologyResponse struct {
	Leader   string           `json:"Leader"`
	NodeTag  string           `json:"NodeTag"`
	Topology *ClusterTopology `json:"Topology"`
	// note: the response returns more info
	// see https://app.quicktype.io?share=pzquGxXJcXyMncfA9JPa for fuller definition
}
