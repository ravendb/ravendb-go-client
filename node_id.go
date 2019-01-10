package ravendb

// NodeID describes a node
type NodeID struct {
	NodeTag         string `json:"NodeTag"`
	NodeURL         string `json:"NodeUrl"`
	ResponsibleNode string `json:"ResponsibleNode"`
}
