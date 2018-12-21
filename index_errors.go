package ravendb

// IndexErrors describes index errors
type IndexErrors struct {
	Name   string           `json:"Name"`
	Errors []*IndexingError `json:"Errors"`
}
