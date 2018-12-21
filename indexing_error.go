package ravendb

// IndexingError describes indexing error message from the server
type IndexingError struct {
	Error     string `json:"Error"`
	Timestamp Time   `json:"Timestamp"`
	Document  string `json:"Document"`
	Action    string `json:"Action"`
}

func (e *IndexingError) String() string {
	return "Error: " + e.Error + ", Document: " + e.Document + ", Action: " + e.Action
}
