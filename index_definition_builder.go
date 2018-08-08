package ravendb

type IndexDefinitionBuilder struct {
	_indexName string

	smap   string // Note: in Go map is a reserved keyword
	reduce string

	storesStrings            map[string]FieldStorage
	indexesStrings           map[string]FieldIndexing
	analyzersStrings         map[string]string
	suggestionsOptions       *StringSet
	termVectorsStrings       map[string]FieldTermVector
	spatialIndexesStrings    map[string]*SpatialOptions
	lockMode                 IndexLockMode
	priority                 IndexPriority
	outputReduceToCollection string
	additionalSources        map[string]string
}

func NewIndexDefinitionBuilder(indexName string) *IndexDefinitionBuilder {
	if indexName == "" {
		indexName = "IndexDefinitionBuilder" // TODO: is it getClass().getSimpleName() ?
	}
	// TODO: make an error
	panicIf(len(indexName) > 256, "The index name is limited to 256 characters, but was: %s", indexName)
	return &IndexDefinitionBuilder{
		_indexName:            indexName,
		storesStrings:         make(map[string]FieldStorage),
		indexesStrings:        make(map[string]FieldIndexing),
		suggestionsOptions:    NewStringSet(),
		analyzersStrings:      make(map[string]string),
		termVectorsStrings:    make(map[string]FieldTermVector),
		spatialIndexesStrings: make(map[string]*SpatialOptions),
	}
}

func (d *IndexDefinitionBuilder) toIndexDefinition(conventions *DocumentConventions, validateMap bool) *IndexDefinition {
	if d.smap == "" && validateMap {
		panicIf(true, "Map is required to generate an index, you cannot create an index without a valid Map property (in index "+d._indexName+").")
		// TODO: return error IllegalStateException("Map is required to generate an index, you cannot create an index without a valid Map property (in index " + _indexName + ").");
	}

	indexDefinition := NewIndexDefinition()
	indexDefinition.setName(d._indexName)
	indexDefinition.setReduce(d.reduce)
	indexDefinition.setLockMode(d.lockMode)
	indexDefinition.setPriority(d.priority)
	indexDefinition.setOutputReduceToCollection(d.outputReduceToCollection)
	indexDefinition.updateIndexType()

	suggestions := make(map[string]bool)
	for _, suggestionsOption := range d.suggestionsOptions.strings {
		suggestions[suggestionsOption] = true
	}

	// TODO: figure out a better way to do it. In Java applyValues() is templated function
	{
		f := func(options *IndexFieldOptions, value FieldIndexing) {
			options.setIndexing(value)
		}
		d.applyFieldIndexingValues(indexDefinition, d.indexesStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value FieldStorage) {
			options.setStorage(value)
		}
		d.applyFieldStorageValues(indexDefinition, d.storesStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value string) {
			options.setAnalyzer(value)
		}
		d.applyStringValues(indexDefinition, d.analyzersStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value FieldTermVector) {
			options.setTermVector(value)
		}
		d.applyFieldTermVectorValues(indexDefinition, d.termVectorsStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value *SpatialOptions) {
			options.setSpatial(value)
		}
		d.applySpatialOptionsValues(indexDefinition, d.spatialIndexesStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value bool) {
			options.setSuggestions(value)
		}
		d.applyBoolValues(indexDefinition, suggestions, f)
	}

	if d.smap != "" {
		indexDefinition.getMaps().add(d.smap)
	}

	indexDefinition.setAdditionalSources(d.additionalSources)
	return indexDefinition
}

func (d *IndexDefinitionBuilder) applyFieldIndexingValues(indexDefinition *IndexDefinition, values map[string]FieldIndexing, action func(*IndexFieldOptions, FieldIndexing)) {
	for key, value := range values {
		fields := indexDefinition.getFields()
		field, ok := fields[key]
		if !ok {
			field = NewIndexFieldOptions()
			fields[key] = field
		}
		action(field, value)
	}
}

func (d *IndexDefinitionBuilder) applyFieldStorageValues(indexDefinition *IndexDefinition, values map[string]FieldStorage, action func(*IndexFieldOptions, FieldStorage)) {
	for key, value := range values {
		fields := indexDefinition.getFields()
		field, ok := fields[key]
		if !ok {
			field = NewIndexFieldOptions()
			fields[key] = field
		}
		action(field, value)
	}
}

func (d *IndexDefinitionBuilder) applyStringValues(indexDefinition *IndexDefinition, values map[string]string, action func(*IndexFieldOptions, string)) {
	for key, value := range values {
		fields := indexDefinition.getFields()
		field, ok := fields[key]
		if !ok {
			field = NewIndexFieldOptions()
			fields[key] = field
		}
		action(field, value)
	}
}

func (d *IndexDefinitionBuilder) applyFieldTermVectorValues(indexDefinition *IndexDefinition, values map[string]FieldTermVector, action func(*IndexFieldOptions, FieldTermVector)) {
	for key, value := range values {
		fields := indexDefinition.getFields()
		field, ok := fields[key]
		if !ok {
			field = NewIndexFieldOptions()
			fields[key] = field
		}
		action(field, value)
	}
}

func (d *IndexDefinitionBuilder) applySpatialOptionsValues(indexDefinition *IndexDefinition, values map[string]*SpatialOptions, action func(*IndexFieldOptions, *SpatialOptions)) {
	for key, value := range values {
		fields := indexDefinition.getFields()
		field, ok := fields[key]
		if !ok {
			field = NewIndexFieldOptions()
			fields[key] = field
		}
		action(field, value)
	}
}

func (d *IndexDefinitionBuilder) applyBoolValues(indexDefinition *IndexDefinition, values map[string]bool, action func(*IndexFieldOptions, bool)) {
	for key, value := range values {
		fields := indexDefinition.getFields()
		field, ok := fields[key]
		if !ok {
			field = NewIndexFieldOptions()
			fields[key] = field
		}
		action(field, value)
	}
}

func (d *IndexDefinitionBuilder) getMap() string {
	return d.smap
}

func (d *IndexDefinitionBuilder) setMap(smap string) {
	d.smap = smap
}

func (d *IndexDefinitionBuilder) getReduce() string {
	return d.reduce
}

func (d *IndexDefinitionBuilder) setReduce(reduce string) {
	d.reduce = reduce
}

func (d *IndexDefinitionBuilder) getStoresStrings() map[string]FieldStorage {
	return d.storesStrings
}

func (d *IndexDefinitionBuilder) setStoresStrings(storesStrings map[string]FieldStorage) {
	d.storesStrings = storesStrings
}

func (d *IndexDefinitionBuilder) getIndexesStrings() map[string]FieldIndexing {
	return d.indexesStrings
}

func (d *IndexDefinitionBuilder) setIndexesStrings(indexesStrings map[string]FieldIndexing) {
	d.indexesStrings = indexesStrings
}

func (d *IndexDefinitionBuilder) getAnalyzersStrings() map[string]string {
	return d.analyzersStrings
}

func (d *IndexDefinitionBuilder) setAnalyzersStrings(analyzersStrings map[string]string) {
	d.analyzersStrings = analyzersStrings
}

func (d *IndexDefinitionBuilder) getSuggestionsOptions() *StringSet {
	return d.suggestionsOptions
}

func (d *IndexDefinitionBuilder) setSuggestionsOptions(suggestionsOptions *StringSet) {
	d.suggestionsOptions = suggestionsOptions
}

func (d *IndexDefinitionBuilder) getTermVectorsStrings() map[string]FieldTermVector {
	return d.termVectorsStrings
}

func (d *IndexDefinitionBuilder) setTermVectorsStrings(termVectorsStrings map[string]FieldTermVector) {
	d.termVectorsStrings = termVectorsStrings
}

func (d *IndexDefinitionBuilder) getSpatialIndexesStrings() map[string]*SpatialOptions {
	return d.spatialIndexesStrings
}

func (d *IndexDefinitionBuilder) setSpatialIndexesStrings(spatialIndexesStrings map[string]*SpatialOptions) {
	d.spatialIndexesStrings = spatialIndexesStrings
}

func (d *IndexDefinitionBuilder) getLockMode() IndexLockMode {
	return d.lockMode
}

func (d *IndexDefinitionBuilder) setLockMode(lockMode IndexLockMode) {
	d.lockMode = lockMode
}

func (d *IndexDefinitionBuilder) getPriority() IndexPriority {
	return d.priority
}

func (d *IndexDefinitionBuilder) setPriority(priority IndexPriority) {
	d.priority = priority
}

func (d *IndexDefinitionBuilder) getOutputReduceToCollection() string {
	return d.outputReduceToCollection
}

func (d *IndexDefinitionBuilder) setOutputReduceToCollection(outputReduceToCollection string) {
	d.outputReduceToCollection = outputReduceToCollection
}

func (d *IndexDefinitionBuilder) getAdditionalSources() map[string]string {
	return d.additionalSources
}

func (d *IndexDefinitionBuilder) setAdditionalSources(additionalSources map[string]string) {
	d.additionalSources = additionalSources
}
