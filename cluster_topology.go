package ravendb

// ClusterTopology is a part of ClusterTopologyResponse
type ClusterTopology struct {
	TopologyID string `json:"TopologyId"`
	AllNodes    map[string]string `json:"AllNodes"`
	// Those map name like A to server url like http://localhost:9999
	Members     map[string]string `json:"Members"`
	Promotables map[string]string `json:"Promotables"`
	Watchers    map[string]string `json:"Watchers"`
	LastNodeId  string            `json:"LastNodeId"`
	Etag        int               `json:"Etag"`
}

// contains returns true if topology contains a given node
func (t *ClusterTopology) contains(node string) bool {
	if t.Members != nil {
		if _, ok := t.Members[node]; ok {
			return true
		}
	}

	if t.Promotables != nil {
		if _, ok := t.Promotables[node]; ok {
			return true
		}
	}

	if t.Watchers != nil {
		if _, ok := t.Watchers[node]; ok {
			return true
		}
	}

	return false
}

func (t *ClusterTopology) getUrlFromTag(tag string) string {
	if tag == "" {
		return ""
	}

	if t.Members != nil {
		if v, ok := t.Members[tag]; ok {
			return v
		}
	}

	if t.Promotables != nil {
		if v, ok := t.Promotables[tag]; ok {
			return v
		}
	}

	if t.Watchers != nil {
		if v, ok := t.Watchers[tag]; ok {
			return v
		}
	}

	return ""
}

// getAllNodes returns all nodes
func (t *ClusterTopology) getAllNodes() map[string]string {
	res := map[string]string{}
	for name, uri := range t.Members {
		res[name] = uri
	}
	for name, uri := range t.Promotables {
		res[name] = uri
	}
	for name, uri := range t.Watchers {
		res[name] = uri
	}
	return res
}
