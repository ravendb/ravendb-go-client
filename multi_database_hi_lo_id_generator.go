package ravendb

import "sync"

// MultiDatabaseHiLoIDGenerator manages per-database HiLoKeyGenerotr
type MultiDatabaseHiLoIDGenerator struct {
	store       *DocumentStore
	conventions *DocumentConventions
	// string -> *MultiTypeHiLoIDGenerator
	_generators sync.Map
}

// NewMultiDatabaseHiLoIDGenerator creates new MultiDatabaseHiLoKeyGenerator
func NewMultiDatabaseHiLoIDGenerator(store *DocumentStore, conventions *DocumentConventions) *MultiDatabaseHiLoIDGenerator {
	return &MultiDatabaseHiLoIDGenerator{
		store:       store,
		conventions: conventions,
	}
}

// GenerateDocumentID generates id
func (g *MultiDatabaseHiLoIDGenerator) GenerateDocumentID(dbName string, entity interface{}) (string, error) {
	if dbName == "" {
		dbName = g.store.database
	}
	panicIf(dbName == "", "expected non-empty dbName")
	generatorI, ok := g._generators.Load(dbName)
	var generator *MultiTypeHiLoIDGenerator
	if !ok {
		generator = NewMultiTypeHiLoIDGenerator(g.store, dbName, g.conventions)
		g._generators.Store(dbName, generator)
	} else {
		generator = generatorI.(*MultiTypeHiLoIDGenerator)
	}
	return generator.GenerateDocumentID(entity)
}

// ReturnUnusedRange returns unused range for all generators
func (g *MultiDatabaseHiLoIDGenerator) ReturnUnusedRange() {
	cb := func(key, value interface{}) bool {
		generator := value.(*MultiTypeHiLoIDGenerator)
		generator.ReturnUnusedRange()
		return true
	}
	g._generators.Range(cb)
}
