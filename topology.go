package ravendb

// Topology describes server nodes
// Result of
// {"Nodes":[{"Url":"http://localhost:9999","ClusterTag":"A","Database":"PyRavenDB","ServerRole":"Rehab"}],"Etag":10}
type Topology struct {
	Nodes []*ServerNode `json:"Nodes"`
	Etag  int           `json:"Etag"`
}

// NewTopology creates a new Topology
func NewTopology() *Topology {
	return &Topology{}
}

func (t *Topology) getEtag() int {
	return t.Etag
}

func (t *Topology) setEtag(etag int) {
	t.Etag = etag
}

func (t *Topology) getNodes() []*ServerNode {
	return t.Nodes
}

func (t *Topology) setNodes(nodes []*ServerNode) {
	t.Nodes = nodes
}
