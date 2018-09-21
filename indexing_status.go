package ravendb

type IndexingStatus struct {
	Status  IndexRunningStatus `json:"Status"`
	Indexes []*IndexStatus     `json:"Indexes"`
}

type IndexStatus struct {
	Name   string             `json:"Name"`
	Status IndexRunningStatus `json:"Status"`
}
