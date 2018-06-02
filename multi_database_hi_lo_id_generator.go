package ravendb

// MultiDatabaseHiLoIDGenerator manages per-database HiLoKeyGenerotr
type MultiDatabaseHiLoIDGenerator struct {
	store      *DocumentStore
	generators map[string]*MultiTypeHiLoIDGenerator
}

// NewMultiDatabaseHiLoIDGenerator creates new MultiDatabaseHiLoKeyGenerator
func NewMultiDatabaseHiLoIDGenerator(store *DocumentStore) *MultiDatabaseHiLoIDGenerator {
	return &MultiDatabaseHiLoIDGenerator{
		store:      store,
		generators: map[string]*MultiTypeHiLoIDGenerator{},
	}
}

// GenerateDocumentID generates id
func (g *MultiDatabaseHiLoIDGenerator) GenerateDocumentID(dbName string, entity interface{}) string {
	if dbName == "" {
		dbName = g.store.database
	}
	panicIf(dbName == "", "expected non-empty dbName")
	generator, ok := g.generators[dbName]
	if !ok {
		generator = NewMultiTypeHiLoIDGenerator(g.store, dbName)
		g.generators[dbName] = generator
	}
	return generator.GenerateDocumentID(entity)
}

// ReturnUnusedRange returns unused range for all generators
func (g *MultiDatabaseHiLoIDGenerator) ReturnUnusedRange() {
	for _, generator := range g.generators {
		generator.ReturnUnusedRange()
	}
}
