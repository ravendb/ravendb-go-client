package ravendb

type IndexingStatus struct {
	Status  IndexRunningStatus `json:"Status"`
	Indexes []*IndexStatus     `json:"Indexes"`
}

func (s *IndexingStatus) getStatus() IndexRunningStatus {
	return s.Status
}

func (s *IndexingStatus) setStatus(status IndexRunningStatus) {
	s.Status = status
}

func (s *IndexingStatus) getIndexes() []*IndexStatus {
	return s.Indexes
}

func (s *IndexingStatus) setIndexes(indexes []*IndexStatus) {
	s.Indexes = indexes
}

type IndexStatus struct {
	Name   string             `json:"Name"`
	Status IndexRunningStatus `json:"Status"`
}

func (s *IndexStatus) getName() string {
	return s.Name
}

func (s *IndexStatus) setName(name string) {
	s.Name = name
}

func (s *IndexStatus) getStatus() IndexRunningStatus {
	return s.Status
}

func (s *IndexStatus) setStatus(status IndexRunningStatus) {
	s.Status = status
}
