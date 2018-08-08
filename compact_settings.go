package ravendb

// CompactSettings is an argument to CompactDatabaseOperation
type CompactSettings struct {
	DatabaseName string   `json:"DatabaseName"`
	Documents    bool     `json:"Documents"`
	Indexes      []string `json:"Indexes,omitempty`
}

// NewCompactSettings creates CompactSettings
func NewCompactSettings() *CompactSettings {
	return &CompactSettings{}
}

func (s *CompactSettings) getDatabaseName() string {
	return s.DatabaseName
}

func (s *CompactSettings) isDocuments() bool {
	return s.Documents
}

func (s *CompactSettings) getIndexes() []string {
	return s.Indexes
}

func (s *CompactSettings) setDatabaseName(databaseName string) {
	s.DatabaseName = databaseName
}

func (s *CompactSettings) setDocuments(documents bool) {
	s.Documents = documents
}

func (s *CompactSettings) setIndexes(indexes []string) {
	s.Indexes = indexes
}
