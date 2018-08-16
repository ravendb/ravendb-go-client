package ravendb

// CompactSettings is an argument to CompactDatabaseOperation
type CompactSettings struct {
	DatabaseName string   `json:"DatabaseName"`
	Documents    bool     `json:"Documents"`
	Indexes      []string `json:"Indexes,omitempty"`
}

// NewCompactSettings creates CompactSettings
func NewCompactSettings() *CompactSettings {
	return &CompactSettings{}
}

func (s *CompactSettings) GetDatabaseName() string {
	return s.DatabaseName
}

func (s *CompactSettings) IsDocuments() bool {
	return s.Documents
}

func (s *CompactSettings) GetIndexes() []string {
	return s.Indexes
}

func (s *CompactSettings) SetDatabaseName(databaseName string) {
	s.DatabaseName = databaseName
}

func (s *CompactSettings) SetDocuments(documents bool) {
	s.Documents = documents
}

func (s *CompactSettings) SetIndexes(indexes []string) {
	s.Indexes = indexes
}
