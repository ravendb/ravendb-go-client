package ravendb

// DatabaseRecord represents database record
type DatabaseRecord struct {
	DatabaseName         string            `json:"DatabaseName"`
	Disabled             bool              `json:"Disabled"`
	DataDirectory        string            `json:"DataDirectory,omitempty"`
	Settings             map[string]string `json:"Settings"`
	ConflictSolverConfig *conflictSolver   `json:"ConflictSolverConfig"`
}

// NewDatabaseRecord returns new database record
func NewDatabaseRecord() *DatabaseRecord {
	return &DatabaseRecord{
		Settings: map[string]string{},
	}
}
