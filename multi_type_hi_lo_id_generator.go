package ravendb

import "sync"

// MultiTypeHiLoIDGenerator manages per-type HiLoKeyGenerator
type MultiTypeHiLoIDGenerator struct {
	_generatorLock sync.Mutex // protects _idGeneratorsByTag
	// maps type name to its generator
	_idGeneratorsByTag map[string]*HiLoIDGenerator
	store              *DocumentStore
	dbName             string
	conventions        *DocumentConventions
}

// NewMultiTypeHiLoIDGenerator creates MultiTypeHiLoKeyGenerator
func NewMultiTypeHiLoIDGenerator(store *DocumentStore, dbName string, conventions *DocumentConventions) *MultiTypeHiLoIDGenerator {
	return &MultiTypeHiLoIDGenerator{
		store:              store,
		dbName:             dbName,
		conventions:        conventions,
		_idGeneratorsByTag: map[string]*HiLoIDGenerator{},
	}
}

// GenerateDocumentID generates a unique key for entity using its type to
// partition keys
func (g *MultiTypeHiLoIDGenerator) GenerateDocumentID(entity interface{}) string {
	typeTagName := g.conventions.GetCollectionName(entity)
	if typeTagName == "" {
		return ""
	}

	tag := g.conventions.GetTransformClassCollectionNameToDocumentIdPrefix()(typeTagName)

	g._generatorLock.Lock()
	value, ok := g._idGeneratorsByTag[tag]
	if !ok {
		value = NewHiLoIdGenerator(tag, g.store, g.dbName, g.conventions.GetIdentityPartsSeparator())
		g._idGeneratorsByTag[tag] = value
	}
	g._generatorLock.Unlock()
	return value.GenerateDocumentID(entity)
}

// ReturnUnusedRange returns unused range for all generators
func (g *MultiTypeHiLoIDGenerator) ReturnUnusedRange() {
	for _, generator := range g._idGeneratorsByTag {
		generator.ReturnUnusedRange()
	}
}
