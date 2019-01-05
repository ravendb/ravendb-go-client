package ravendb

// GetDocumentsResult is a result of GetDocument command
type GetDocumentsResult struct {
	Includes      map[string]interface{}   `json:"Includes"`
	Results       []map[string]interface{} `json:"Results"`
	NextPageStart int                      `json:"NextPageStart"`
}
