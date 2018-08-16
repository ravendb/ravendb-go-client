package ravendb

type ConflictSolver struct {
	ResolveByCollection map[string]*ScriptResolver `json:"ResolveByCollection"`
	ResolveToLatest     bool                       `json:"ResolveToLatest"`
}

func NewConflictSolver() *ConflictSolver {
	return &ConflictSolver{}
}

func (s *ConflictSolver) getResolveByCollection() map[string]*ScriptResolver {
	return s.ResolveByCollection
}

func (s *ConflictSolver) setResolveByCollection(resolveByCollection map[string]*ScriptResolver) {
	s.ResolveByCollection = resolveByCollection
}

func (s *ConflictSolver) isResolveToLatest() bool {
	return s.ResolveToLatest
}

func (s *ConflictSolver) setResolveToLatest(resolveToLatest bool) {
	s.ResolveToLatest = resolveToLatest
}
