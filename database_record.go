package ravendb

type DatabaseRecord struct {
	DatabaseName         string            `json:"DatabaseName"`
	Disabled             bool              `json:"Disabled"`
	DataDirectory        *string           `json:"DataDirectory"`
	Settings             map[string]string `json:"Settings"`
	ConflictSolverConfig *conflictSolver   `json:"ConflictSolverConfig"`
}

func NewDatabaseRecord() *DatabaseRecord {
	return &DatabaseRecord{
		Settings: map[string]string{},
	}
}
