package ravendb

type IndexDefinition struct {
	Name              string                        `json:"Name"`
	Priority          *IndexPriority                `json:"Priority"`
	LockMode          *IndexLockMode                `json:"LockMode"`
	AdditionalSources map[string]string             `json:"AdditionalSources"`
	Maps              []string                      `json:"Maps"`
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
	return res
}

func (d *IndexDefinition) GetPriority() *IndexPriority {
	return d.Priority
}

func (d *IndexDefinition) SetPriority(priority IndexPriority) {
	d.Priority = toStrPtr(priority)
}

func (d *IndexDefinition) GetLockMode() *IndexLockMode {
	return d.LockMode
}

func (d *IndexDefinition) SetLockMode(lockMode IndexLockMode) {
	d.LockMode = toStrPtr(lockMode)
}

func (d *IndexDefinition) GetAdditionalSources() map[string]string {
	if d.AdditionalSources == nil {
		d.AdditionalSources = make(map[string]string)
	}
	return d.AdditionalSources
}

func (d *IndexDefinition) SetAdditionalSources(additionalSources map[string]string) {
	// preserve additionalSources being always non-nil
	// so that JSON serializes it as {} and not 'null'
	if additionalSources == nil {
		if len(d.AdditionalSources) == 0 {
			return
		}
		additionalSources = make(map[string]string)
	}
	d.AdditionalSources = additionalSources
}

func (d *IndexDefinition) String() string {
	return d.Name
}

func (d *IndexDefinition) GetFields() map[string]*IndexFieldOptions {
	if d.Fields == nil {
		d.Fields = make(map[string]*IndexFieldOptions)
	}
	return d.Fields
}

func (d *IndexDefinition) GetConfiguration() IndexConfiguration {
	if d.Configuration == nil {
		d.Configuration = NewIndexConfiguration()
	}
	return d.Configuration
}

func (d *IndexDefinition) SetConfiguration(configuration IndexConfiguration) {
	d.Configuration = configuration
}

// Note: this must be called after finishing building index definition to set IndexType
// In Java it's calculated on demand via getType
func (d *IndexDefinition) updateIndexTypeAndMaps() {
	if d.IndexType == "" || d.IndexType == IndexType_NONE {
		d.IndexType = d.detectStaticIndexType()
	}
	d.Maps = StringArrayRemoveDuplicates(d.Maps)
}

func (d *IndexDefinition) GetType() IndexType {
	if d.IndexType == "" || d.IndexType == IndexType_NONE {
		d.IndexType = d.detectStaticIndexType()
	}

	return d.IndexType
}

func (d *IndexDefinition) SetType(indexType IndexType) {
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

func (d *IndexDefinition) GetOutputReduceToCollection() *string {
	return d.OutputReduceToCollection
}

func (d *IndexDefinition) SetOutputReduceToCollection(outputReduceToCollection string) {
	d.OutputReduceToCollection = toStrPtr(outputReduceToCollection)
}
