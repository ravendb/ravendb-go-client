package ravendb

// ConflictSolver describes how to resolve conflicts
type ConflictSolver struct {
	ResolveByCollection map[string]*ScriptResolver `json:"ResolveByCollection"`
	ResolveToLatest     bool                       `json:"ResolveToLatest"`
}
