package ravendb

type IndexingStatus struct {
	Status  IndexRunningStatus `json:"Status"`
	Indexes []*IndexStatus     `json:"Indexes"`
}

func (s *IndexingStatus) GetStatus() IndexRunningStatus {
	return s.Status
}

func (s *IndexingStatus) SetStatus(status IndexRunningStatus) {
	s.Status = status
}

func (s *IndexingStatus) GetIndexes() []*IndexStatus {
	return s.Indexes
}

func (s *IndexingStatus) SetIndexes(indexes []*IndexStatus) {
	s.Indexes = indexes
}

type IndexStatus struct {
	Name   string             `json:"Name"`
	Status IndexRunningStatus `json:"Status"`
}

func (s *IndexStatus) GetName() string {
	return s.Name
}

func (s *IndexStatus) SetName(name string) {
	s.Name = name
}

func (s *IndexStatus) GetStatus() IndexRunningStatus {
	return s.Status
}

func (s *IndexStatus) SetStatus(status IndexRunningStatus) {
	s.Status = status
}
