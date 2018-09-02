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

func (c *WktCriteria) ToQueryToken(fieldName string, addQueryParameter func(Object) string) queryToken {
	return c.SpatialCriteriaCommon.toQueryTokenCommon(c, fieldName, addQueryParameter)
}

func (c *WktCriteria) GetShapeToken(addQueryParameter func(Object) string) *shapeToken {
	return ShapeToken_wkt(addQueryParameter(c._shapeWkt))
}
