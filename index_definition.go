package ravendb

type IndexDefinition struct {
	name              string                        `json:"Name"`
	priority          IndexPriority                 `json:"Priority"`
	lockMode          IndexLockMode                 `json:"LockMode"`
	additionalSources map[string]string             `json:"AdditionalSources"`
	maps              *StringSet                    `json:"Maps"`
	reduce            string                        `json:"Reduce"`
	fields            map[string]*IndexFieldOptions `json:"Fields"`
	configuration     IndexConfiguration            `json:"Configuration"`
	indexType         IndexType                     `json:"IndexType"`
	//TBD 4.1  bool testIndex;
	outputReduceToCollection string `json:"OutputReduceToCollection"`
}

func NewIndexDefinition() *IndexDefinition {
	return &IndexDefinition{
		configuration: NewIndexConfiguration(),
	}
}

func (d *IndexDefinition) getName() string {
	return d.name
}

func (d *IndexDefinition) setName(name string) {
	d.name = name
}

func (d *IndexDefinition) getPriority() IndexPriority {
	return d.priority
}

func (d *IndexDefinition) setPriority(priority IndexPriority) {
	d.priority = priority
}

func (d *IndexDefinition) getLockMode() IndexLockMode {
	return d.lockMode
}

func (d *IndexDefinition) setLockMode(lockMode IndexLockMode) {
	d.lockMode = lockMode
}

func (d *IndexDefinition) getAdditionalSources() map[string]string {
	if d.additionalSources == nil {
		d.additionalSources = make(map[string]string)
	}
	return d.additionalSources
}

func (d *IndexDefinition) setAdditionalSources(additionalSources map[string]string) {
	d.additionalSources = additionalSources
}

func (d *IndexDefinition) getMaps() *StringSet {
	if d.maps == nil {
		d.maps = NewStringSet()
	}
	return d.maps
}

func (d *IndexDefinition) setMaps(maps *StringSet) {
	d.maps = maps
}

func (d *IndexDefinition) getReduce() string {
	return d.reduce
}

func (d *IndexDefinition) setReduce(reduce string) {
	d.reduce = reduce
}

func (d *IndexDefinition) String() string {
	return d.name
}

func (d *IndexDefinition) getFields() map[string]*IndexFieldOptions {
	if d.fields == nil {
		d.fields = make(map[string]*IndexFieldOptions)
	}
	return d.fields
}

func (d *IndexDefinition) setFields(fields map[string]*IndexFieldOptions) {
	d.fields = fields
}

func (d *IndexDefinition) getConfiguration() IndexConfiguration {
	if d.configuration == nil {
		d.configuration = NewIndexConfiguration()
	}
	return d.configuration
}

func (d *IndexDefinition) setConfiguration(configuration IndexConfiguration) {
	d.configuration = configuration
}

func (d *IndexDefinition) getType() IndexType {
	if d.indexType == "" || d.indexType == IndexType_NONE {
		d.indexType = d.detectStaticIndexType()
	}

	return d.indexType
}

func (d *IndexDefinition) setType(indexType IndexType) {
	d.indexType = indexType
}

func (d *IndexDefinition) detectStaticIndexType() IndexType {
	if StringUtils_isBlank(d.reduce) {
		return IndexType_MAP
	}
	return IndexType_MAP_REDUCE
}

//TBD 4.1  bool isTestIndex()

//TBD 4.1   setTestIndex(bool testIndex)

func (d *IndexDefinition) getOutputReduceToCollection() string {
	return d.outputReduceToCollection
}

func (d *IndexDefinition) setOutputReduceToCollection(outputReduceToCollection string) {
	d.outputReduceToCollection = outputReduceToCollection
}
