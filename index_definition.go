package ravendb

type IndexDefinition struct {
	Name              string                        `json:"Name"`
	Priority          *IndexPriority                `json:"Priority"`
	LockMode          *IndexLockMode                `json:"LockMode"`
	AdditionalSources map[string]string             `json:"AdditionalSources"`
	Maps              *StringSet                    `json:"Maps"`
	Reduce            *string                       `json:"Reduce"`
	Fields            map[string]*IndexFieldOptions `json:"Fields"`
	Configuration     IndexConfiguration            `json:"Configuration"`
	IndexType         IndexType                     `json:"Type"`
	//TBD 4.1  bool testIndex;
	OutputReduceToCollection *string `json:"OutputReduceToCollection"`
}

func toStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func NewIndexDefinition() *IndexDefinition {
	res := &IndexDefinition{
		Configuration: NewIndexConfiguration(),

		// Note: initializing those is possibly wasteful but
		// it's needed to serialize them like Java, as {} and not null
		Fields:            make(map[string]*IndexFieldOptions),
		AdditionalSources: make(map[string]string),
	}
	// force setting the right index type
	// Note: in Java it's calculated on demand
	res.getType()
	return res
}

func (d *IndexDefinition) getName() string {
	return d.Name
}

func (d *IndexDefinition) setName(name string) {
	d.Name = name
}

func (d *IndexDefinition) getPriority() *IndexPriority {
	return d.Priority
}

func (d *IndexDefinition) setPriority(priority IndexPriority) {
	d.Priority = toStrPtr(priority)
}

func (d *IndexDefinition) getLockMode() *IndexLockMode {
	return d.LockMode
}

func (d *IndexDefinition) setLockMode(lockMode IndexLockMode) {
	d.LockMode = toStrPtr(lockMode)
}

func (d *IndexDefinition) getAdditionalSources() map[string]string {
	if d.AdditionalSources == nil {
		d.AdditionalSources = make(map[string]string)
	}
	return d.AdditionalSources
}

func (d *IndexDefinition) setAdditionalSources(additionalSources map[string]string) {
	// preserve additionalSources being always non-nil
	// to JSON serialize as {} and not nil
	if additionalSources == nil {
		if len(d.AdditionalSources) == 0 {
			return
		}
		additionalSources = make(map[string]string)
	}
	d.AdditionalSources = additionalSources
}

func (d *IndexDefinition) getMaps() *StringSet {
	if d.Maps == nil {
		d.Maps = NewStringSet()
	}
	return d.Maps
}

func (d *IndexDefinition) setMaps(maps *StringSet) {
	d.Maps = maps
}

func (d *IndexDefinition) getReduce() *string {
	return d.Reduce
}

func (d *IndexDefinition) setReduce(reduce string) {
	d.Reduce = toStrPtr(reduce)
}

func (d *IndexDefinition) String() string {
	return d.Name
}

func (d *IndexDefinition) getFields() map[string]*IndexFieldOptions {
	if d.Fields == nil {
		d.Fields = make(map[string]*IndexFieldOptions)
	}
	return d.Fields
}

func (d *IndexDefinition) setFields(fields map[string]*IndexFieldOptions) {
	d.Fields = fields
}

func (d *IndexDefinition) getConfiguration() IndexConfiguration {
	if d.Configuration == nil {
		d.Configuration = NewIndexConfiguration()
	}
	return d.Configuration
}

func (d *IndexDefinition) setConfiguration(configuration IndexConfiguration) {
	d.Configuration = configuration
}

func (d *IndexDefinition) getType() IndexType {
	if d.IndexType == "" || d.IndexType == IndexType_NONE {
		d.IndexType = d.detectStaticIndexType()
	}

	return d.IndexType
}

func (d *IndexDefinition) setType(indexType IndexType) {
	if indexType == "" {
		indexType = IndexType_NONE
	}
	d.IndexType = indexType
}

func (d *IndexDefinition) detectStaticIndexType() IndexType {
	if d.Reduce == nil || StringUtils_isBlank(*d.Reduce) {
		return IndexType_MAP
	}
	return IndexType_MAP_REDUCE
}

//TBD 4.1  bool isTestIndex()

//TBD 4.1   setTestIndex(bool testIndex)

func (d *IndexDefinition) getOutputReduceToCollection() *string {
	return d.OutputReduceToCollection
}

func (d *IndexDefinition) setOutputReduceToCollection(outputReduceToCollection string) {
	d.OutputReduceToCollection = toStrPtr(outputReduceToCollection)
}
