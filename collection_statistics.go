package ravendb

// CollectionStatistics describes collection statistics
type CollectionStatistics struct {
	CountOfDocuments int            `json:"CountOfDocuments"`
	CountOfConflicts int            `json:"CountOfConflicts"`
	Collections      map[string]int `json:"Collections"`
}
