package ravendb

var (
	SpatialCriteriaFactory_INSTANCE = NewSpatialCriteriaFactory()
)

type SpatialCriteriaFactory struct {
}

func NewSpatialCriteriaFactory() *SpatialCriteriaFactory {
	return &SpatialCriteriaFactory{}
}

func (f *SpatialCriteriaFactory) relatesToShape(shapeWkt string, relation SpatialRelation) *WktCriteria {
	return f.relatesToShapeWithError(shapeWkt, relation, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
}

func (f *SpatialCriteriaFactory) relatesToShapeWithError(shapeWkt string, relation SpatialRelation, distErrorPercent float64) *WktCriteria {
	return NewWktCriteria(shapeWkt, relation, distErrorPercent)
}

func (f *SpatialCriteriaFactory) intersects(shapeWkt string) *WktCriteria {
	return f.intersectsWithError(shapeWkt, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
}

func (f *SpatialCriteriaFactory) intersectsWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.relatesToShapeWithError(shapeWkt, SpatialRelation_INTERSECTS, distErrorPercent)
}

func (f *SpatialCriteriaFactory) contains(shapeWkt string) *WktCriteria {
	return f.containsWithError(shapeWkt, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
}

func (f *SpatialCriteriaFactory) containsWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.relatesToShapeWithError(shapeWkt, SpatialRelation_CONTAINS, distErrorPercent)
}

func (f *SpatialCriteriaFactory) disjoint(shapeWkt string) *WktCriteria {
	return f.disjointWithError(shapeWkt, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
}

func (f *SpatialCriteriaFactory) disjointWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.relatesToShapeWithError(shapeWkt, SpatialRelation_DISJOINT, distErrorPercent)
}

func (f *SpatialCriteriaFactory) within(shapeWkt string) *WktCriteria {
	return f.withinWithError(shapeWkt, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
}

func (f *SpatialCriteriaFactory) withinWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.relatesToShapeWithError(shapeWkt, SpatialRelation_WITHIN, distErrorPercent)
}

func (f *SpatialCriteriaFactory) withinRadius(radius float64, latitude float64, longitude float64) *CircleCriteria {
	return f.withinRadiusWithUnits(radius, latitude, longitude, "")
}

func (f *SpatialCriteriaFactory) withinRadiusWithUnits(radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits) *CircleCriteria {
	return f.withinRadiusWithUnitsAndError(radius, latitude, longitude, radiusUnits, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
}

func (f *SpatialCriteriaFactory) withinRadiusWithUnitsAndError(radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distErrorPercent float64) *CircleCriteria {
	return NewCircleCriteria(radius, latitude, longitude, radiusUnits, SpatialRelation_WITHIN, distErrorPercent)
}
