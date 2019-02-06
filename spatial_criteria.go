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

	var whereOperator whereOperator

	switch c._relation {
	case SpatialRelationWithin:
		whereOperator = whereOperatorSpatialWithin
	case SpatialRelationContains:
		whereOperator = whereOperatorSpatialContains
	case SpatialRelationDisjoin:
		whereOperator = whereOperatorSpatialDisjoint
	case SpatialRelationIntersects:
		whereOperator = whereOperatorSpatialIntersects
	default:
		//throw new IllegalArgumentError();
		panicIf(true, "Unknown relation '%s'", c._relation)
	}

	opts := newWhereOptionsWithTokenAndDistance(shapeToken, c._distanceErrorPct)
	return createWhereTokenWithOptions(whereOperator, fieldName, "", opts)
}
