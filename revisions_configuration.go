package ravendb

type RevisionsConfiguration struct {
	DefaultConfig *RevisionsCollectionConfiguration `json:"Default"`

	Collections map[string]*RevisionsCollectionConfiguration `json:"Collections"`
}
