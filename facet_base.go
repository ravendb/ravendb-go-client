package ravendb

type FacetBase interface {
	// those are supplied by each type
	ToFacetToken(addQueryParameter func(interface{}) string) *facetToken

	// inherited from FacetBaseCommon
	GetOptions() *FacetOptions
	GetAggregations() map[FacetAggregation]string
	SetDisplayFieldName(string)
	SetOptions(*FacetOptions)
}

type FacetBaseCommon struct {
	displayFieldName string
	options          *FacetOptions
	aggregations     map[FacetAggregation]string
}

func NewFacetBaseCommon() FacetBaseCommon {
	return FacetBaseCommon{
		aggregations: make(map[FacetAggregation]string),
	}
}

func (f *FacetBaseCommon) GetDisplayFieldName() string {
	return f.displayFieldName
}

func (f *FacetBaseCommon) SetDisplayFieldName(displayFieldName string) {
	f.displayFieldName = displayFieldName
}

func (f *FacetBaseCommon) GetOptions() *FacetOptions {
	return f.options
}

func (f *FacetBaseCommon) SetOptions(options *FacetOptions) {
	f.options = options
}

func (f *FacetBaseCommon) GetAggregations() map[FacetAggregation]string {
	return f.aggregations
}

func (f *FacetBaseCommon) SetAggregations(aggregations map[FacetAggregation]string) {
	f.aggregations = aggregations
}
