package ravendb

// GetDocumentsResult is a result of GetDocument command
type GetDocumentsResult struct {
	Includes      map[string]interface{} `json:"Includes"`
	Results       ArrayNode              `json:"Results"`
	NextPageStart int                    `json:"NextPageStart"`
}
