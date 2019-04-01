package ravendb

type WktCriteria struct {
	SpatialCriteriaCommon
	_shapeWkt    string
	_radiusUnits SpatialUnits
}

func NewWktCriteria(shapeWkt string, relation SpatialRelation, radiusUnits SpatialUnits, distErrorPercent float64) *WktCriteria {
	res := &WktCriteria{
		_shapeWkt:    shapeWkt,
		_radiusUnits: radiusUnits,
	}
	res._relation = relation
	res._distanceErrorPct = distErrorPercent
	return res
}

func (c *WktCriteria) ToQueryToken(fieldName string, addQueryParameter func(interface{}) string) queryToken {
	return c.SpatialCriteriaCommon.toQueryTokenCommon(c, fieldName, addQueryParameter)
}

func (c *WktCriteria) GetShapeToken(addQueryParameter func(interface{}) string) *shapeToken {
	return shapeTokenWkt(addQueryParameter(c._shapeWkt), c._radiusUnits)
}
