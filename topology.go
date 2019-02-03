package ravendb

// Topology describes server nodes
type Topology struct {
	Nodes []*ServerNode `json:"Nodes"`
	Etag  int64         `json:"Etag"`
}
