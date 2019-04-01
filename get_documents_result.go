package ravendb

// GetDocumentsResult is a result of GetDocument command
type GetDocumentsResult struct {
	Includes        map[string]interface{}   `json:"Includes"`
	Results         []map[string]interface{} `json:"Results"`
	CounterIncludes map[string]interface{}   `json:"CounterIncludes"`
	NextPageStart   int                      `json:"NextPageStart"`
}
