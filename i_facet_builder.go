package ravendb

type IFacetBuilder interface {
	byRanges(rang *RangeBuilder, ranges ...*RangeBuilder) IFacetOperations

	byField(fieldName string) IFacetOperations

	allResults() IFacetOperations

	//TBD expr IFacetOperations<T> ByField(Expression<Func<T, object>> path);
	//TBD expr IFacetOperations<T> ByRanges(Expression<Func<T, bool>> path, params Expression<Func<T, bool>>[] paths);
}
