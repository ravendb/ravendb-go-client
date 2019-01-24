package ravendb

type FacetBase interface {
	// those are supplied by each type
	// TODO: make it private
	ToFacetToken(addQueryParameter func(interface{}) string) *facetToken

	// those are inherited from FacetBaseCommon
	GetOptions() *FacetOptions
	GetAggregations() map[FacetAggregation]string
	SetDisplayFieldName(string)
	SetOptions(*FacetOptions)
}

type FacetBaseCommon struct {
	DisplayFieldName string                      `json:"DisplayFieldName,omitempty"`
	Options          *FacetOptions               `json:"Options"`
	Aggregations     map[FacetAggregation]string `json:"Aggregations"`
}

func NewFacetBaseCommon() FacetBaseCommon {
	return FacetBaseCommon{
		Aggregations: make(map[FacetAggregation]string),
	}
}

func (f *FacetBaseCommon) GetDisplayFieldName() string {
	return f.DisplayFieldName
}

func (f *FacetBaseCommon) SetDisplayFieldName(displayFieldName string) {
	f.DisplayFieldName = displayFieldName
}

func (f *FacetBaseCommon) GetOptions() *FacetOptions {
	return f.Options
}

func (f *FacetBaseCommon) SetOptions(options *FacetOptions) {
	f.Options = options
}

func (f *FacetBaseCommon) GetAggregations() map[FacetAggregation]string {
	return f.Aggregations
}

func (f *FacetBaseCommon) SetAggregations(aggregations map[FacetAggregation]string) {
	f.Aggregations = aggregations
}
