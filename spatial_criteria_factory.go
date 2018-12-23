package ravendb

var (
	SpatialCriteriaFactory_INSTANCE = NewSpatialCriteriaFactory()
)

type SpatialCriteriaFactory struct {
}

func NewSpatialCriteriaFactory() *SpatialCriteriaFactory {
	return &SpatialCriteriaFactory{}
}

func (f *SpatialCriteriaFactory) RelatesToShape(shapeWkt string, relation SpatialRelation) *WktCriteria {
	return f.RelatesToShapeWithError(shapeWkt, relation, IndexingSpatialDefaultDistnaceErrorPct)
}

func (f *SpatialCriteriaFactory) RelatesToShapeWithError(shapeWkt string, relation SpatialRelation, distErrorPercent float64) *WktCriteria {
	return NewWktCriteria(shapeWkt, relation, distErrorPercent)
}

func (f *SpatialCriteriaFactory) Intersects(shapeWkt string) *WktCriteria {
	return f.IntersectsWithError(shapeWkt, IndexingSpatialDefaultDistnaceErrorPct)
}

func (f *SpatialCriteriaFactory) IntersectsWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.RelatesToShapeWithError(shapeWkt, SpatialRelation_INTERSECTS, distErrorPercent)
}

func (f *SpatialCriteriaFactory) Contains(shapeWkt string) *WktCriteria {
	return f.ContainsWithError(shapeWkt, IndexingSpatialDefaultDistnaceErrorPct)
}

func (f *SpatialCriteriaFactory) ContainsWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.RelatesToShapeWithError(shapeWkt, SpatialRelation_CONTAINS, distErrorPercent)
}

func (f *SpatialCriteriaFactory) Disjoint(shapeWkt string) *WktCriteria {
	return f.DisjointWithError(shapeWkt, IndexingSpatialDefaultDistnaceErrorPct)
}

func (f *SpatialCriteriaFactory) DisjointWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.RelatesToShapeWithError(shapeWkt, SpatialRelation_DISJOINT, distErrorPercent)
}

func (f *SpatialCriteriaFactory) Within(shapeWkt string) *WktCriteria {
	return f.WithinWithError(shapeWkt, IndexingSpatialDefaultDistnaceErrorPct)
}

func (f *SpatialCriteriaFactory) WithinWithError(shapeWkt string, distErrorPercent float64) *WktCriteria {
	return f.RelatesToShapeWithError(shapeWkt, SpatialRelation_WITHIN, distErrorPercent)
}

func (f *SpatialCriteriaFactory) WithinRadius(radius float64, latitude float64, longitude float64) *CircleCriteria {
	return f.WithinRadiusWithUnits(radius, latitude, longitude, "")
}

func (f *SpatialCriteriaFactory) WithinRadiusWithUnits(radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits) *CircleCriteria {
	return f.WithinRadiusWithUnitsAndError(radius, latitude, longitude, radiusUnits, IndexingSpatialDefaultDistnaceErrorPct)
}

func (f *SpatialCriteriaFactory) WithinRadiusWithUnitsAndError(radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distErrorPercent float64) *CircleCriteria {
	return NewCircleCriteria(radius, latitude, longitude, radiusUnits, SpatialRelation_WITHIN, distErrorPercent)
}
