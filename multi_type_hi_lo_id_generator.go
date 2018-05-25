package ravendb

import "sync"

// MultiTypeHiLoIDGenerator manages per-type HiLoKeyGenerator
type MultiTypeHiLoIDGenerator struct {
	store  *DocumentStore
	dbName string
	// maps type name to its generator
	_idGeneratorsByTag map[string]*HiLoIDGenerator
	lock               sync.Mutex // protects _idGeneratorsByTag
	// TODO: conventions
}

// NewMultiTypeHiLoIDGenerator creates MultiTypeHiLoKeyGenerator
func NewMultiTypeHiLoIDGenerator(store *DocumentStore, dbName string) *MultiTypeHiLoIDGenerator {
	return &MultiTypeHiLoIDGenerator{
		store:              store,
		dbName:             dbName,
		_idGeneratorsByTag: map[string]*HiLoIDGenerator{},
	}
}

// GenerateDocumentID generates a unique key for entity using its type to
// partition keys
func (g *MultiTypeHiLoIDGenerator) GenerateDocumentID(entity interface{}) string {
	tag := defaultTransformTypeTagName(getShortTypeName(entity))
	g.lock.Lock()
	generator, ok := g._idGeneratorsByTag[tag]
	if !ok {
		generator = NewHiLoIDGenerator(tag, g.store, g.dbName)
		g._idGeneratorsByTag[tag] = generator
	}
	g.lock.Unlock()
	return generator.GenerateDocumentID()
}

// ReturnUnusedRange returns unused range for all generators
func (g *MultiTypeHiLoIDGenerator) ReturnUnusedRange() {
	for _, generator := range g._idGeneratorsByTag {
		generator.ReturnUnusedRange()
	}
}
