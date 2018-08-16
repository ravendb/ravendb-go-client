package ravendb

type CollectionStatistics struct {
	CountOfDocuments int            `json:"CountOfDocuments"`
	CountOfConflicts int            `json:"CountOfConflicts"`
	Collections      map[string]int `json:"Collections"`
}

func NewCollectionStatistics() *CollectionStatistics {
	return &CollectionStatistics{}
}

func (s *CollectionStatistics) GetCollections() map[string]int {
	return s.Collections
}

func (s *CollectionStatistics) SetCollections(collections map[string]int) {
	s.Collections = collections
}

func (s *CollectionStatistics) GetCountOfDocuments() int {
	return s.CountOfDocuments
}

func (s *CollectionStatistics) SetCountOfDocuments(countOfDocuments int) {
	s.CountOfDocuments = countOfDocuments
}

func (s *CollectionStatistics) GetCountOfConflicts() int {
	return s.CountOfConflicts
}

func (s *CollectionStatistics) SetCountOfConflicts(countOfConflicts int) {
	s.CountOfConflicts = countOfConflicts
}
