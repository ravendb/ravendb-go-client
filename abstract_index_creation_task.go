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

func (t *AbstractIndexCreationTask) GetAdditionalSources() map[string]string {
	return t.AdditionalSources
}

func (t *AbstractIndexCreationTask) SetAdditionalSources(additionalSources map[string]string) {
	t.AdditionalSources = additionalSources
}

func (t *AbstractIndexCreationTask) CreateIndexDefinition() *IndexDefinition {
	if t.Conventions == nil {
		t.Conventions = NewDocumentConventions()
	}

	indexDefinitionBuilder := NewIndexDefinitionBuilder(t.GetIndexName())
	indexDefinitionBuilder.indexesStrings = t.IndexesStrings
	indexDefinitionBuilder.analyzersStrings = t.AnalyzersStrings
	indexDefinitionBuilder.setMap(t.Map)
	indexDefinitionBuilder.reduce = t.Reduce
	indexDefinitionBuilder.storesStrings = t.StoresStrings
	indexDefinitionBuilder.suggestionsOptions = t.IndexSuggestions
	indexDefinitionBuilder.termVectorsStrings = t.TermVectorsStrings
	indexDefinitionBuilder.spatialIndexesStrings = t.SpatialOptionsStrings
	indexDefinitionBuilder.outputReduceToCollection = t.OutputReduceToCollection
	indexDefinitionBuilder.additionalSources = t.GetAdditionalSources()

	return indexDefinitionBuilder.toIndexDefinition(t.Conventions, false)
}

func (t *AbstractIndexCreationTask) IsMapReduce() bool {
	return t.Reduce != ""
}

func (t *AbstractIndexCreationTask) GetIndexName() string {
	panicIf(t.IndexName == "", "indexName must be set by 'sub-class' to be equivalent of Java's getClass().getSimpleName()")
	return strings.Replace(t.IndexName, "_", "/", -1)
}

func (t *AbstractIndexCreationTask) GetConventions() *DocumentConventions {
	return t.Conventions
}

func (t *AbstractIndexCreationTask) SetConventions(conventions *DocumentConventions) {
	t.Conventions = conventions
}

func (t *AbstractIndexCreationTask) GetPriority() IndexPriority {
	return t.Priority
}

func (t *AbstractIndexCreationTask) SetPriority(priority IndexPriority) {
	t.Priority = priority
}

func (t *AbstractIndexCreationTask) GetLockMode() IndexLockMode {
	return t.LockMode
}

func (t *AbstractIndexCreationTask) SetLockMode(lockMode IndexLockMode) {
	t.LockMode = lockMode
}

func (t *AbstractIndexCreationTask) Execute(store *IDocumentStore) error {
	return store.ExecuteIndex(t)
}

func (t *AbstractIndexCreationTask) Execute2(store *IDocumentStore, conventions *DocumentConventions, database string) error {
	return t.PutIndex(store, conventions, database)
}

func (t *AbstractIndexCreationTask) PutIndex(store *IDocumentStore, conventions *DocumentConventions, database string) error {
	oldConventions := t.GetConventions()
	defer t.SetConventions(oldConventions)

	conv := conventions
	if conv == nil {
		conv = t.GetConventions()
	}
	if conv == nil {
		conv = store.GetConventions()
	}
	t.SetConventions(conv)

	indexDefinition := t.CreateIndexDefinition()
	indexDefinition.Name = t.GetIndexName()

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

func (t *AbstractIndexCreationTask) Index(field string, indexing FieldIndexing) {
	t.IndexesStrings[field] = indexing
}

func (t *AbstractIndexCreationTask) Spatial(field string, indexing func(*SpatialOptionsFactory) *SpatialOptions) {
	v := indexing(NewSpatialOptionsFactory())
	t.SpatialOptionsStrings[field] = v
}

func (t *AbstractIndexCreationTask) StoreAllFields(storage FieldStorage) {
	t.StoresStrings[Constants_Documents_Indexing_Fields_ALL_FIELDS] = storage
}

func (t *AbstractIndexCreationTask) Store(field string, storage FieldStorage) {
	t.StoresStrings[field] = storage
}

func (t *AbstractIndexCreationTask) Analyze(field string, analyzer string) {
	t.AnalyzersStrings[field] = analyzer
}

func (t *AbstractIndexCreationTask) TermVector(field string, termVector FieldTermVector) {
	t.TermVectorsStrings[field] = termVector
}

func (t *AbstractIndexCreationTask) Suggestion(field string) {
	t.IndexSuggestions.Add(field)
}
