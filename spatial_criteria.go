package ravendb

type SpatialCriteria interface {
	getShapeToken(addQueryParameter func(Object) string) *ShapeToken
	toQueryToken(fieldName string, addQueryParameter func(Object) string) QueryToken
}

type SpatialCriteriaCommon struct {
	_relation         SpatialRelation
	_distanceErrorPct float64
}

func NewSpatialCriteria(relation SpatialRelation, distanceErrorPct float64) SpatialCriteriaCommon {
	return SpatialCriteriaCommon{
		_relation:         relation,
		_distanceErrorPct: distanceErrorPct,
	}
}

// Note: hacky way to emulate Java's inheritance
func (c *SpatialCriteriaCommon) toQueryTokenCommon(sc SpatialCriteria, fieldName string, addQueryParameter func(Object) string) QueryToken {
	shapeToken := sc.getShapeToken(addQueryParameter)

	var whereOperator WhereOperator

	switch c._relation {
	case SpatialRelation_WITHIN:
		whereOperator = WhereOperator_SPATIAL_WITHIN
	case SpatialRelation_CONTAINS:
		whereOperator = WhereOperator_SPATIAL_CONTAINS
	case SpatialRelation_DISJOINT:
		whereOperator = WhereOperator_SPATIAL_DISJOINT
	case SpatialRelation_INTERSECTS:
		whereOperator = WhereOperator_SPATIAL_INTERSECTS
	default:
		//throw new IllegalArgumentException();
		panicIf(true, "Unknown relation '%s'", c._relation)
	}

	opts := NewWhereOptionsWithTokenAndDistance(shapeToken, c._distanceErrorPct)
	return WhereToken_createWithOptions(whereOperator, fieldName, "", opts)
}
