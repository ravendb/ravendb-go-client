package ravendb

type ConflictSolver struct {
	resolveByCollection map[string]*ScriptResolver
	resolveToLatest     bool
}

func NewConflictSolver() *ConflictSolver {
	return &ConflictSolver{}
}

func (s *ConflictSolver) getResolveByCollection() map[string]*ScriptResolver {
	return s.resolveByCollection
}

func (s *ConflictSolver) setResolveByCollection(resolveByCollection map[string]*ScriptResolver) {
	s.resolveByCollection = resolveByCollection
}

func (s *ConflictSolver) isResolveToLatest() bool {
	return s.resolveToLatest
}

func (s *ConflictSolver) setResolveToLatest(resolveToLatest bool) {
	s.resolveToLatest = resolveToLatest
}
