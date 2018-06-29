package ravendb

type AbstractIndexCreationTask struct {
	smap   string // Note: in Go map is a reserved keyword
	reduce string

	conventions       *DocumentConventions
	additionalSources map[string]string
	priority          IndexPriority
	lockMode          IndexLockMode

	storesStrings         map[string]FieldStorage
	indexesStrings        map[string]FieldIndexing
	analyzersStrings      map[string]string
	indexSuggestions      *StringSet
	termVectorsStrings    map[string]FieldTermVector
	spatialOptionsStrings map[string]*SpatialOptions

	outputReduceToCollection string

	// in Go this must be set by "sub-class". In Java it's dynamically calculated
	// as getClass().getSimpleName()
	indexName string
}

// Note: in Java we subclass AbstractIndexCreationTask and indexName is derived
// from derived class name. In Go we don't subclass and must provide index name
// manually
func NewAbstractIndexCreationTask(indexName string) *AbstractIndexCreationTask {
	panicIf(indexName == "", "indexName cannot be empty")
	return &AbstractIndexCreationTask{
		storesStrings:         make(map[string]FieldStorage),
		indexesStrings:        make(map[string]FieldIndexing),
		analyzersStrings:      make(map[string]string),
		indexSuggestions:      NewStringSet(),
		termVectorsStrings:    make(map[string]FieldTermVector),
		spatialOptionsStrings: make(map[string]*SpatialOptions),

		indexName: indexName,
	}
}

func (t *AbstractIndexCreationTask) getAdditionalSources() map[string]string {
	return t.additionalSources
}

func (t *AbstractIndexCreationTask) setAdditionalSources(additionalSources map[string]string) {
	t.additionalSources = additionalSources
}

func (t *AbstractIndexCreationTask) createIndexDefinition() *IndexDefinition {
	if t.conventions == nil {
		t.conventions = NewDocumentConventions()
	}

	indexDefinitionBuilder := NewIndexDefinitionBuilder(t.getIndexName())
	indexDefinitionBuilder.setIndexesStrings(t.indexesStrings)
	indexDefinitionBuilder.setAnalyzersStrings(t.analyzersStrings)
	indexDefinitionBuilder.setMap(t.smap)
	indexDefinitionBuilder.setReduce(t.reduce)
	indexDefinitionBuilder.setStoresStrings(t.storesStrings)
	indexDefinitionBuilder.setSuggestionsOptions(t.indexSuggestions)
	indexDefinitionBuilder.setTermVectorsStrings(t.termVectorsStrings)
	indexDefinitionBuilder.setSpatialIndexesStrings(t.spatialOptionsStrings)
	indexDefinitionBuilder.setOutputReduceToCollection(t.outputReduceToCollection)
	indexDefinitionBuilder.setAdditionalSources(t.getAdditionalSources())

	return indexDefinitionBuilder.toIndexDefinition(t.conventions, false)
}

func (t *AbstractIndexCreationTask) isMapReduce() bool {
	return t.reduce != ""
}

func (t *AbstractIndexCreationTask) getIndexName() string {
	panicIf(t.indexName == "", "indexName must be set by 'sub-class' to be equivalent of Java's getClass().getSimpleName() with '_' replaced with '/'")
	return t.indexName
}

func (t *AbstractIndexCreationTask) getConventions() *DocumentConventions {
	return t.conventions
}

func (t *AbstractIndexCreationTask) setConventions(conventions *DocumentConventions) {
	t.conventions = conventions
}

func (t *AbstractIndexCreationTask) getPriority() IndexPriority {
	return t.priority
}

func (t *AbstractIndexCreationTask) setPriority(priority IndexPriority) {
	t.priority = priority
}

func (t *AbstractIndexCreationTask) getLockMode() IndexLockMode {
	return t.lockMode
}

func (t *AbstractIndexCreationTask) setLockMode(lockMode IndexLockMode) {
	t.lockMode = lockMode
}

func (t *AbstractIndexCreationTask) execute(store *IDocumentStore) error {
	return store.executeIndex(t)
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
		conv = store.getConventions()
	}
	t.setConventions(conv)

	indexDefinition := t.createIndexDefinition()
	indexDefinition.setName(t.getIndexName())

	if t.lockMode != "" {
		indexDefinition.setLockMode(t.lockMode)
	}

	if t.priority != "" {
		indexDefinition.setPriority(t.priority)
	}

	op := NewPutIndexesOperation(indexDefinition)
	if database == "" {
		database = store.getDatabase()
	}
	return store.maintenance().forDatabase(database).send(op)
}

func (t *AbstractIndexCreationTask) index(field string, indexing FieldIndexing) {
	t.indexesStrings[field] = indexing
}

func (t *AbstractIndexCreationTask) spatial(field string, indexing func(*SpatialOptionsFactory) *SpatialOptions) {
	v := indexing(NewSpatialOptionsFactory())
	t.spatialOptionsStrings[field] = v
}

func (t *AbstractIndexCreationTask) storeAllFields(storage FieldStorage) {
	t.storesStrings[Constants_Documents_Indexing_Fields_ALL_FIELDS] = storage
}

func (t *AbstractIndexCreationTask) store(field string, storage FieldStorage) {
	t.storesStrings[field] = storage
}

func (t *AbstractIndexCreationTask) analyze(field string, analyzer string) {
	t.analyzersStrings[field] = analyzer
}

func (t *AbstractIndexCreationTask) termVector(field string, termVector FieldTermVector) {
	t.termVectorsStrings[field] = termVector
}

func (t *AbstractIndexCreationTask) suggestion(field string) {
	t.indexSuggestions.add(field)
}
