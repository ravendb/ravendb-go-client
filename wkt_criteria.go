package ravendb

type WktCriteria struct {
	SpatialCriteriaCommon
	_shapeWkt string
}

func NewWktCriteria(shapeWkt string, relation SpatialRelation, distErrorPercent float64) *WktCriteria {
	res := &WktCriteria{
		_shapeWkt: shapeWkt,
	}
	res._relation = relation
	res._distanceErrorPct = distErrorPercent
	return res
}

func (c *WktCriteria) toQueryToken(fieldName string, addQueryParameter func(Object) string) QueryToken {
	return c.SpatialCriteriaCommon.toQueryTokenCommon(c, fieldName, addQueryParameter)
}

func (c *WktCriteria) getShapeToken(addQueryParameter func(Object) string) *ShapeToken {
	return ShapeToken_wkt(addQueryParameter(c._shapeWkt))
}
