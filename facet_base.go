package ravendb

type FacetBase interface {
	// those are supplied by each type
	ToFacetToken(addQueryParameter func(interface{}) string) (*facetToken, error)

	// those are inherited from FacetBaseCommon
	SetDisplayFieldName(string)
	GetOptions() *FacetOptions
	SetOptions(*FacetOptions)
	GetAggregations() map[FacetAggregation]string
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
