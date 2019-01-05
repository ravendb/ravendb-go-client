package ravendb

type SpatialCriteria interface {
	GetShapeToken(addQueryParameter func(interface{}) string) *shapeToken
	ToQueryToken(fieldName string, addQueryParameter func(interface{}) string) queryToken
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
func (c *SpatialCriteriaCommon) toQueryTokenCommon(sc SpatialCriteria, fieldName string, addQueryParameter func(interface{}) string) queryToken {
	shapeToken := sc.GetShapeToken(addQueryParameter)

	var whereOperator WhereOperator

	switch c._relation {
	case SpatialRelationWithin:
		whereOperator = WhereOperatorSpatialWithin
	case SpatialRelationContains:
		whereOperator = WhereOperatorSpatialContains
	case SpatialRelationDisjoin:
		whereOperator = WhereOperatorSpatialDisjoint
	case SpatialRelationIntersects:
		whereOperator = WhereOperatorSpatialIntersects
	default:
		//throw new IllegalArgumentError();
		panicIf(true, "Unknown relation '%s'", c._relation)
	}

	opts := NewWhereOptionsWithTokenAndDistance(shapeToken, c._distanceErrorPct)
	return createWhereTokenWithOptions(whereOperator, fieldName, "", opts)
}
