package ravendb

import "strings"

// AbstractIndexCreationTask is for creating an index
// TODO: rename to IndexCreationTask
type AbstractIndexCreationTask struct {
	Map    string // Note: in Go map is a reserved keyword
	Reduce string

	Conventions       *DocumentConventions
	AdditionalSources map[string]string
	Priority          IndexPriority
	LockMode          IndexLockMode

	StoresStrings         map[string]FieldStorage
	IndexesStrings        map[string]FieldIndexing
	AnalyzersStrings      map[string]string
	IndexSuggestions      *StringSet
	TermVectorsStrings    map[string]FieldTermVector
	SpatialOptionsStrings map[string]*SpatialOptions

	OutputReduceToCollection string

	// in Go this must be set by "sub-class". In Java it's dynamically calculated
	// as getClass().getSimpleName()
	IndexName string
}

// NewAbstractIndexCreationTask creates AbstractIndexCreationTask
// Note: in Java we subclass AbstractIndexCreationTask and indexName is derived
// from derived class name. In Go we don't subclass and must provide index name
// manually
func NewAbstractIndexCreationTask(indexName string) *AbstractIndexCreationTask {
	panicIf(indexName == "", "indexName cannot be empty")
	return &AbstractIndexCreationTask{
		StoresStrings:         make(map[string]FieldStorage),
		IndexesStrings:        make(map[string]FieldIndexing),
		AnalyzersStrings:      make(map[string]string),
		IndexSuggestions:      NewStringSet(),
		TermVectorsStrings:    make(map[string]FieldTermVector),
		SpatialOptionsStrings: make(map[string]*SpatialOptions),

		IndexName: indexName,
	}
}

func (t *AbstractIndexCreationTask) getAdditionalSources() map[string]string {
	return t.AdditionalSources
}

func (t *AbstractIndexCreationTask) setAdditionalSources(additionalSources map[string]string) {
	t.AdditionalSources = additionalSources
}

func (t *AbstractIndexCreationTask) createIndexDefinition() *IndexDefinition {
	if t.Conventions == nil {
		t.Conventions = NewDocumentConventions()
	}

	indexDefinitionBuilder := NewIndexDefinitionBuilder(t.getIndexName())
	indexDefinitionBuilder.setIndexesStrings(t.IndexesStrings)
	indexDefinitionBuilder.setAnalyzersStrings(t.AnalyzersStrings)
	indexDefinitionBuilder.setMap(t.Map)
	indexDefinitionBuilder.setReduce(t.Reduce)
	indexDefinitionBuilder.setStoresStrings(t.StoresStrings)
	indexDefinitionBuilder.setSuggestionsOptions(t.IndexSuggestions)
	indexDefinitionBuilder.setTermVectorsStrings(t.TermVectorsStrings)
	indexDefinitionBuilder.setSpatialIndexesStrings(t.SpatialOptionsStrings)
	indexDefinitionBuilder.setOutputReduceToCollection(t.OutputReduceToCollection)
	indexDefinitionBuilder.setAdditionalSources(t.getAdditionalSources())

	return indexDefinitionBuilder.toIndexDefinition(t.Conventions, false)
}

func (t *AbstractIndexCreationTask) isMapReduce() bool {
	return t.Reduce != ""
}

func (t *AbstractIndexCreationTask) getIndexName() string {
	panicIf(t.IndexName == "", "indexName must be set by 'sub-class' to be equivalent of Java's getClass().getSimpleName()")
	return strings.Replace(t.IndexName, "_", "/", -1)
}

func (t *AbstractIndexCreationTask) getConventions() *DocumentConventions {
	return t.Conventions
}

func (t *AbstractIndexCreationTask) setConventions(conventions *DocumentConventions) {
	t.Conventions = conventions
}

func (t *AbstractIndexCreationTask) getPriority() IndexPriority {
	return t.Priority
}

func (t *AbstractIndexCreationTask) setPriority(priority IndexPriority) {
	t.Priority = priority
}

func (t *AbstractIndexCreationTask) getLockMode() IndexLockMode {
	return t.LockMode
}

func (t *AbstractIndexCreationTask) setLockMode(lockMode IndexLockMode) {
	t.LockMode = lockMode
}

func (t *AbstractIndexCreationTask) execute(store *IDocumentStore) error {
	return store.ExecuteIndex(t)
}

func (t *AbstractIndexCreationTask) execute2(store *IDocumentStore, conventions *DocumentConventions, database string) error {
	return t.putIndex(store, conventions, database)
}

func (t *AbstractIndexCreationTask) putIndex(store *IDocumentStore, conventions *DocumentConventions, database string) error {
	oldConventions := t.getConventions()
	defer t.setConventions(oldConventions)

	conv := conventions
	if conv == nil {
		conv = t.getConventions()
	}
	if conv == nil {
		conv = store.GetConventions()
	}
	t.setConventions(conv)

	indexDefinition := t.createIndexDefinition()
	indexDefinition.SetName(t.getIndexName())

	if t.LockMode != "" {
		indexDefinition.SetLockMode(t.LockMode)
	}

	if t.Priority != "" {
		indexDefinition.SetPriority(t.Priority)
	}

	op := NewPutIndexesOperation(indexDefinition)
	if database == "" {
		database = store.GetDatabase()
	}
	return store.Maintenance().ForDatabase(database).Send(op)
}

func (t *AbstractIndexCreationTask) index(field string, indexing FieldIndexing) {
	t.IndexesStrings[field] = indexing
}

func (t *AbstractIndexCreationTask) spatial(field string, indexing func(*SpatialOptionsFactory) *SpatialOptions) {
	v := indexing(NewSpatialOptionsFactory())
	t.SpatialOptionsStrings[field] = v
}

func (t *AbstractIndexCreationTask) storeAllFields(storage FieldStorage) {
	t.StoresStrings[Constants_Documents_Indexing_Fields_ALL_FIELDS] = storage
}

func (t *AbstractIndexCreationTask) store(field string, storage FieldStorage) {
	t.StoresStrings[field] = storage
}

func (t *AbstractIndexCreationTask) analyze(field string, analyzer string) {
	t.AnalyzersStrings[field] = analyzer
}

func (t *AbstractIndexCreationTask) termVector(field string, termVector FieldTermVector) {
	t.TermVectorsStrings[field] = termVector
}

func (t *AbstractIndexCreationTask) suggestion(field string) {
	t.IndexSuggestions.add(field)
}
