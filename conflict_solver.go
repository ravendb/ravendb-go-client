package ravendb

type conflictSolver struct {
	ResolveByCollection map[string]*scriptResolver `json:"ResolveByCollection"`
	ResolveToLatest     bool                       `json:"ResolveToLatest"`
}
