package ravendb

type PutIndexesResponse struct {
	Results []*PutIndexResult `json:"Results"`
}
