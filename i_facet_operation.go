package ravendb

type IFacetOperations interface {
	withDisplayName(displayName string) IFacetOperations

	withOptions(options *FacetOptions) IFacetOperations

	sumOn(path string) IFacetOperations
	minOn(path string) IFacetOperations
	maxOn(path string) IFacetOperations
	averageOn(path string) IFacetOperations

	//TBD expr overloads with expression
}
