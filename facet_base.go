package ravendb

type FacetBase interface {
	// those are supplied by each type
	toFacetToken(addQueryParameter func(Object) string) *FacetToken

	// inherited from FacetBaseCommon
	getOptions() *FacetOptions
	getAggregations() map[FacetAggregation]string
	setDisplayFieldName(string)
	setOptions(*FacetOptions)
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

func (f *FacetBaseCommon) getDisplayFieldName() string {
	return f.displayFieldName
}

func (f *FacetBaseCommon) setDisplayFieldName(displayFieldName string) {
	f.displayFieldName = displayFieldName
}

func (f *FacetBaseCommon) getOptions() *FacetOptions {
	return f.options
}

func (f *FacetBaseCommon) setOptions(options *FacetOptions) {
	f.options = options
}

func (f *FacetBaseCommon) getAggregations() map[FacetAggregation]string {
	return f.aggregations
}

func (f *FacetBaseCommon) setAggregations(aggregations map[FacetAggregation]string) {
	f.aggregations = aggregations
}
