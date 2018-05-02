package data

type IndexLockMode string
type IndexPriority string
type SortOptions string
type FieldIndexing string
type FieldTermVector string

const INDEX_LOCK_MODE_UNLOCK = "Unlock"
const INDEX_LOCK_MODE_LOCKED_IGNORE = "LockedIgnore"
const INDEX_LOCK_MODE_LOCKED_ERROR = "LockedError"
const INDEX_LOCK_MODE_SIDE_BY_SIDE = "SideBySide"

const INDEX_PRIORITY_LOW = "Low"
const INDEX_PRIORITY_NORMAL = "Normal"
const INDEX_PRIORITY_HIGH = "High"

const SORT_OPTIONS_NONE = "None"
const SORT_OPTIONS_STRING = "String"
const SORT_OPTIONS_NUMERIC = "Numeric"

const FIELD_INDEXING_NO = "No"
const FIELD_INDEXING_SEARCH = "Search"
const FIELD_INDEXING_EXACT = "Exact"
const FIELD_INDEXING_DEFAULT = "Default"

const FIELD_TERM_VECTOR_NO = "No"
const FIELD_TERM_VECTOR_YES = "Yes"
const FIELD_TERM_VECTOR_WITH_POSITIONS = "WithPositions"
const FIELD_TERM_VECTOR_WITH_OFFSETS = "WithOffsets"
const FIELD_TERM_VECTORWITH_POSITIONS_AND_OFFSETS = "WithPositionsAndOffsets"

type IndexDefinition struct {
	Name          string
	Configuration []string
	Reduce        bool
	IndexId       int
	IsTestIndex   bool
	LockMode      IndexLockMode
	Priority      IndexPriority
	Maps          []string
	Fields        map[string]IndexFieldOptions
}

func NewIndexDefinition(name string, maps []string, configuration []string, reduce bool, index_id int, is_test_index bool, lock_mod IndexLockMode, priority IndexPriority, fields map[string]IndexFieldOptions) (*IndexDefinition, error) {
	ref := &IndexDefinition{}

	ref.Name = name
	ref.Configuration = configuration
	ref.Reduce = reduce
	ref.IndexId = index_id
	ref.IsTestIndex = is_test_index
	ref.LockMode = lock_mod
	ref.Priority = priority
	ref.Maps = maps
	ref.Fields = fields

	return ref, nil
}

type IndexFieldOptions struct {
	SortOptions     SortOptions
	FieldIndexing   FieldIndexing
	Storage         bool
	Suggestions     bool
	FieldTermVector FieldTermVector
	Analyzer        string
}

func NewIndexFieldOptions(sort_options SortOptions, field_indexing FieldIndexing, storage bool, suggestions bool, term_vector FieldTermVector, analyzer string) (*IndexFieldOptions, error) {

	ref := &IndexFieldOptions{}
	ref.SortOptions = sort_options
	ref.FieldIndexing = field_indexing
	ref.Storage = storage
	ref.Suggestions = suggestions
	ref.FieldTermVector = term_vector
	ref.Analyzer = analyzer

	return ref, nil
}
func (obj IndexFieldOptions) ToJson() map[string]interface{} {

	var data map[string]interface{}
	data["Analyzer"] = obj.Analyzer
	if obj.FieldIndexing == "" {
		data["Indexing"] = nil
	} else {
		data["Indexing"] = obj.FieldIndexing
	}
	if obj.FieldIndexing == "" {
		data["Sort"] = nil
	} else {
		data["Sort"] = obj.SortOptions
	}
	data["Spatial"] = nil
	if obj.Storage {
		data["Storage"] = "Yes"
	} else {
		data["Storage"] = "No"
	}
	data["Suggestions"] = obj.Suggestions
	if obj.FieldIndexing == "" {
		data["TermVector"] = nil
	} else {
		data["TermVector"] = obj.FieldTermVector
	}

	return data
}
