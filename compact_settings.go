package ravendb

// CompactSettings is an argument to CompactDatabaseOperation
type CompactSettings struct {
	DatabaseName string   `json:"DatabaseName"`
	Documents    bool     `json:"Documents"`
	Indexes      []string `json:"Indexes,omitempty"`
}
