package ravendb

// RevisionsConfiguration describes revisions configuration
type RevisionsConfiguration struct {
	DefaultConfig *RevisionsCollectionConfiguration `json:"Default"`

	Collections map[string]*RevisionsCollectionConfiguration `json:"Collections"`
}
