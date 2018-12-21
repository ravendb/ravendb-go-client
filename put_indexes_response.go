package ravendb

// PutIndexesResponse represents server's response to PutIndexesCommand
type PutIndexesResponse struct {
	Results []*PutIndexResult `json:"Results"`
}
