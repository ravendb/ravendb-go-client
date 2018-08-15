package ravendb

type IFacetBuilder interface {
	ByRanges(rang *RangeBuilder, ranges ...*RangeBuilder) IFacetOperations

	ByField(fieldName string) IFacetOperations

	AllResults() IFacetOperations

	//TBD expr IFacetOperations<T> ByField(Expression<Func<T, object>> path);
	//TBD expr IFacetOperations<T> ByRanges(Expression<Func<T, bool>> path, params Expression<Func<T, bool>>[] paths);
}
