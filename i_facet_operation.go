package ravendb

type IFacetOperations interface {
	WithDisplayName(displayName string) IFacetOperations

	WithOptions(options *FacetOptions) IFacetOperations

	SumOn(path string) IFacetOperations
	MinOn(path string) IFacetOperations
	MaxOn(path string) IFacetOperations
	AverageOn(path string) IFacetOperations

	//TBD expr overloads with expression
}
