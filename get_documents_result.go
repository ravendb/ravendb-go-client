package ravendb

// GetDocumentsResult is a result of GetDocument command
type GetDocumentsResult struct {
	Includes      ObjectNode `json:"Includes"`
	Results       ArrayNode  `json:"Results"`
	NextPageStart int        `json:"NextPageStart"`
}
