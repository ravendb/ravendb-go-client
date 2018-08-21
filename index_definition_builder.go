package ravendb

type IndexDefinitionBuilder struct {
	_indexName string

	smap   string // Note: in Go map is a reserved keyword
	reduce string

	storesStrings            map[string]FieldStorage
	indexesStrings           map[string]FieldIndexing
	analyzersStrings         map[string]string
	suggestionsOptions       []string
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
	indexDefinition.Name = d._indexName
	if d.reduce != "" {
		indexDefinition.Reduce = &d.reduce
	}
	indexDefinition.LockMode = d.lockMode
	indexDefinition.Priority = d.priority
	indexDefinition.SetOutputReduceToCollection(d.outputReduceToCollection)
	indexDefinition.updateIndexTypeAndMaps()

	suggestions := make(map[string]bool)
	for _, suggestionsOption := range d.suggestionsOptions {
		suggestions[suggestionsOption] = true
	}

	// TODO: figure out a better way to do it. In Java applyValues() is templated function
	{
		f := func(options *IndexFieldOptions, value FieldIndexing) {
			options.Indexing = value
		}
		d.applyFieldIndexingValues(indexDefinition, d.indexesStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value FieldStorage) {
			options.Storage = value
		}
		d.applyFieldStorageValues(indexDefinition, d.storesStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value string) {
			options.Analyzer = value
		}
		d.applyStringValues(indexDefinition, d.analyzersStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value FieldTermVector) {
			options.TermVector = value
		}
		d.applyFieldTermVectorValues(indexDefinition, d.termVectorsStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value *SpatialOptions) {
			options.Spatial = value
		}
		d.applySpatialOptionsValues(indexDefinition, d.spatialIndexesStrings, f)
	}

	{
		f := func(options *IndexFieldOptions, value bool) {
			options.Suggestions = value
		}
		d.applyBoolValues(indexDefinition, suggestions, f)
	}

	if d.smap != "" {
		indexDefinition.Maps = append(indexDefinition.Maps, d.smap)
	}

	indexDefinition.SetAdditionalSources(d.additionalSources)
	return indexDefinition
}

func (d *IndexDefinitionBuilder) applyFieldIndexingValues(indexDefinition *IndexDefinition, values map[string]FieldIndexing, action func(*IndexFieldOptions, FieldIndexing)) {
	for key, value := range values {
		fields := indexDefinition.GetFields()
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
		fields := indexDefinition.GetFields()
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
		fields := indexDefinition.GetFields()
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
		fields := indexDefinition.GetFields()
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
		fields := indexDefinition.GetFields()
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
		fields := indexDefinition.GetFields()
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
