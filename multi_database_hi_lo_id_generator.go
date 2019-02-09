package ravendb

// MultiDatabaseHiLoIDGenerator manages per-database HiLoKeyGenerotr
type MultiDatabaseHiLoIDGenerator struct {
	store       *DocumentStore
	conventions *DocumentConventions
	_generators map[string]*MultiTypeHiLoIDGenerator
}

// NewMultiDatabaseHiLoIDGenerator creates new MultiDatabaseHiLoKeyGenerator
func NewMultiDatabaseHiLoIDGenerator(store *DocumentStore, conventions *DocumentConventions) *MultiDatabaseHiLoIDGenerator {
	return &MultiDatabaseHiLoIDGenerator{
		store:       store,
		conventions: conventions,
		_generators: map[string]*MultiTypeHiLoIDGenerator{},
	}
}

// GenerateDocumentID generates id
func (g *MultiDatabaseHiLoIDGenerator) GenerateDocumentID(dbName string, entity interface{}) (string, error) {
	if dbName == "" {
		dbName = g.store.database
	}
	panicIf(dbName == "", "expected non-empty dbName")
	generator, ok := g._generators[dbName]
	if !ok {
		generator = NewMultiTypeHiLoIDGenerator(g.store, dbName, g.conventions)
		g._generators[dbName] = generator
	}
	return generator.GenerateDocumentID(entity)
}

// ReturnUnusedRange returns unused range for all generators
func (g *MultiDatabaseHiLoIDGenerator) ReturnUnusedRange() {
	for _, generator := range g._generators {
		generator.ReturnUnusedRange()
	}
}
