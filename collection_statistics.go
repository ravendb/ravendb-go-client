package ravendb

type CollectionStatistics struct {
	CountOfDocuments int            `json:"CountOfDocuments"`
	CountOfConflicts int            `json:"CountOfConflicts"`
	Collections      map[string]int `json:"Collections"`
}

func NewCollectionStatistics() *CollectionStatistics {
	return &CollectionStatistics{}
}

func (s *CollectionStatistics) getCollections() map[string]int {
	return s.Collections
}

func (s *CollectionStatistics) setCollections(collections map[string]int) {
	s.Collections = collections
}

func (s *CollectionStatistics) getCountOfDocuments() int {
	return s.CountOfDocuments
}

func (s *CollectionStatistics) setCountOfDocuments(countOfDocuments int) {
	s.CountOfDocuments = countOfDocuments
}

func (s *CollectionStatistics) getCountOfConflicts() int {
	return s.CountOfConflicts
}

func (s *CollectionStatistics) setCountOfConflicts(countOfConflicts int) {
	s.CountOfConflicts = countOfConflicts
}
