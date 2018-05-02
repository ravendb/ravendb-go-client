package data

//@param str query: Actual query that will be performed.
//@param dict query_parameters: Parameters to the query
//@param int start:  Number of records that should be skipped.
//@param int page_size:  Maximum number of records that will be retrieved.
//@param bool wait_for_non_stale_results: True to wait in the server side to non stale result
//@param None or float cutoff_etag: Gets or sets the cutoff etag.
type BaseIndexQuery struct {
	PageSize                      uint
	PageSizeIsSet                 bool
	Query                         string
	QueryParams                   []string
	Start                         uint
	WaitForNonStaleResults        bool
	WaitForNonStaleResultsTimeout uint
	CutoffEtag                    float64
}

func NewBaseIndexQuery(query string, query_parameters []string, start uint, page_size uint, wait_for_non_stale_results bool, wait_for_non_stale_results_timeout uint, cutoff_etag float64) (*BaseIndexQuery, error) {

	ref := &BaseIndexQuery{}
	ref.Query = query
	ref.QueryParams = query_parameters
	ref.Start = start
	ref.PageSize = page_size
	ref.WaitForNonStaleResults = wait_for_non_stale_results
	ref.WaitForNonStaleResultsTimeout = wait_for_non_stale_results_timeout
	ref.CutoffEtag = cutoff_etag

	return ref, nil
}
func (obj BaseIndexQuery) getPageSize() uint {
	return obj.PageSize
}
func (obj BaseIndexQuery) getPageSizeSet() bool {
	return obj.PageSizeIsSet
}

type IndexQuery struct {
	BaseIndexQuery                                              *BaseIndexQuery
	AllowMultipleIndexEntriesForSameDocumentToResultTransformer bool
	Includes                                                    []string
	ShowTimings                                                 bool
	SkipDuplicateChecking                                       bool
}

func NewIndexQuery(query string, query_params []string, start uint, includes []string, show_timings bool, skip_duplicate_checking bool, page_size uint, wait_for_non_stale_results bool, wait_for_non_stale_results_timeout uint, cutoff_etag float64) (*IndexQuery, error) {

	ref := &IndexQuery{}
	baseIndexQuery, _ := NewBaseIndexQuery(query, query_params, start, page_size, wait_for_non_stale_results, wait_for_non_stale_results_timeout, cutoff_etag)
	ref.BaseIndexQuery = baseIndexQuery
	ref.Includes = includes
	ref.ShowTimings = show_timings
	ref.SkipDuplicateChecking = skip_duplicate_checking

	return ref, nil
}
func (obj IndexQuery) GetQueryHash() string {
	//todo find xxhash lib with digest
	return ""
}
func (obj IndexQuery) ToJson() map[string]interface{} {
	var data map[string]interface{}
	data["Query"] = obj.BaseIndexQuery.Query
	data["CutoffEtag"] = obj.BaseIndexQuery.CutoffEtag
	if obj.GetPageSizeSet() {
		data["PageSize"] = obj.GetPageSize()
	}
	if obj.BaseIndexQuery.WaitForNonStaleResults {
		data["WaitForNonStaleResults"] = obj.BaseIndexQuery.WaitForNonStaleResults
	}
	if obj.BaseIndexQuery.Start != 0 {
		data["Start"] = obj.BaseIndexQuery.Start
	}
	if obj.BaseIndexQuery.WaitForNonStaleResultsTimeout > 0 {
		data["WaitForNonStaleResultsTimeout"] = obj.BaseIndexQuery.WaitForNonStaleResultsTimeout
	}
	if obj.AllowMultipleIndexEntriesForSameDocumentToResultTransformer {
		data["AllowMultipleIndexEntriesForSameDocumentToResultTransformer"] = obj.AllowMultipleIndexEntriesForSameDocumentToResultTransformer
	}
	if obj.ShowTimings {
		data["ShowTimings"] = obj.ShowTimings
	}
	if obj.SkipDuplicateChecking {
		data["SkipDuplicateChecking"] = obj.SkipDuplicateChecking
	}
	if len(obj.Includes) > 0 {
		data["Includes"] = obj.Includes
	}
	if len(obj.BaseIndexQuery.QueryParams) > 0 {
		data["QueryParameters"] = obj.BaseIndexQuery.QueryParams
	} else {
		data["QueryParameters"] = nil
	}
	return data
}
func (obj IndexQuery) GetPageSize() uint {
	return obj.BaseIndexQuery.getPageSize()
}
func (obj IndexQuery) GetPageSizeSet() bool {
	return obj.BaseIndexQuery.getPageSizeSet()
}
func (obj IndexQuery) GetCustomQueryStrVariables() string {
	return ""
}

type FacetQuery struct {
	BaseIndexQuery *BaseIndexQuery
	Facets         []Facet //todo implement Facet
	FacetSetupDoc  string
}

func NewFacetQuery(query string, facet_setup_doc string, facets []Facet, start uint, page_size uint) (*FacetQuery, error) {

	ref := &FacetQuery{}
	baseIndexQuery, _ := NewBaseIndexQuery(query, []string{}, start, page_size, false, 0, 0)
	ref.BaseIndexQuery = baseIndexQuery
	ref.Facets = facets
	ref.FacetSetupDoc = facet_setup_doc

	return ref, nil
}
func (obj FacetQuery) GetPageSize() uint {
	return obj.BaseIndexQuery.getPageSize()
}
func (obj FacetQuery) GetPageSizeSet() bool {
	return obj.BaseIndexQuery.getPageSizeSet()
}
func (obj FacetQuery) ToJson() map[string]interface{} {
	var data map[string]interface{}
	data["Query"] = obj.BaseIndexQuery.Query
	data["CutoffEtag"] = obj.BaseIndexQuery.CutoffEtag
	if obj.GetPageSizeSet() {
		data["PageSize"] = obj.GetPageSize()
	}
	if obj.BaseIndexQuery.WaitForNonStaleResults {
		data["WaitForNonStaleResults"] = obj.BaseIndexQuery.WaitForNonStaleResults
	}
	if obj.BaseIndexQuery.Start != 0 {
		data["Start"] = obj.BaseIndexQuery.Start
	}
	if obj.BaseIndexQuery.WaitForNonStaleResultsTimeout > 0 {
		data["WaitForNonStaleResultsTimeout"] = obj.BaseIndexQuery.WaitForNonStaleResultsTimeout
	}
	if len(obj.BaseIndexQuery.QueryParams) > 0 {
		data["QueryParameters"] = obj.BaseIndexQuery.QueryParams
	} else {
		data["QueryParameters"] = nil
	}
	if len(obj.Facets) > 0 {
		data["Facets"] = obj.Facets
	}
	if obj.FacetSetupDoc != "" {
		data["FacetSetupDoc"] = obj.FacetSetupDoc
	}
	return data
}
func (obj FacetQuery) GetQueryHash() string {
	//todo find xxhash lib with digest
	return ""
}

type FacetMode string
type FacetAggregation string
type FacetTermSortMode string

const FACET_MODE_DEFAULT FacetMode = "Default"
const FACET_MODE_RANGES FacetMode = "Ranges"

const FACET_AGGREGATION_NONE FacetAggregation = "None"
const FACET_AGGREGATION_COUNT FacetAggregation = "Count"
const FACET_AGGREGATION_MAX FacetAggregation = "Max"
const FACET_AGGREGATION_MIN FacetAggregation = "Min"
const FACET_AGGREGATION_AVERAGE FacetAggregation = "Average"
const FACET_AGGREGATION_SUM FacetAggregation = "Sum"

const FACET_TERM_SORT_MODE_VALUE_ASC FacetTermSortMode = "ValueAsc"
const FACET_TERM_SORT_MODE_VALUE_DESC FacetTermSortMode = "ValueDesc"
const FACET_TERM_SORT_MODE_HITS_ASC FacetTermSortMode = "HitsAsc"
const FACET_TERM_SORT_MODE_HITS_DESC FacetTermSortMode = "HitsDesc"

type Facet struct {
	Name                  string
	DisplayName           string
	Ranges                []int
	Mode                  FacetMode
	Aggregation           FacetAggregation
	AggregationField      string
	AggregationType       string
	MaxResult             int
	TermSortMode          FacetTermSortMode
	IncludeRemainingTerms bool
}

func NewFacet(name string, display_name string, ranges []int, mode FacetMode, aggregation FacetAggregation, aggregation_field string, aggregation_type string, max_result int, term_sort_mode FacetTermSortMode, include_remaining_terms bool) (*Facet, error) {

	ref := &Facet{}
	ref.Name = name
	ref.DisplayName = display_name
	ref.Ranges = ranges
	ref.Mode = mode
	ref.Aggregation = aggregation
	ref.AggregationField = aggregation_field
	ref.AggregationType = aggregation_type
	ref.MaxResult = max_result
	ref.TermSortMode = term_sort_mode
	ref.IncludeRemainingTerms = include_remaining_terms

	return ref, nil
}
func (obj Facet) ToJson() map[string]interface{} {

	var data map[string]interface{}
	data["Mode"] = obj.Mode
	data["Aggregation"] = obj.Aggregation
	data["AggregationField"] = obj.AggregationField
	data["AggregationType"] = obj.AggregationType
	data["Name"] = obj.Name
	data["TermSortMode"] = obj.TermSortMode
	data["IncludeRemainingTerms"] = obj.IncludeRemainingTerms
	if obj.MaxResult > 0 {
		data["MaxResult"] = obj.MaxResult
	}
	if obj.DisplayName != "" && obj.DisplayName != obj.Name {
		data["DisplayName"] = obj.DisplayName
	}
	if len(obj.Ranges) > 0 {
		data["Ranges"] = obj.Ranges
	}

	return data
}
