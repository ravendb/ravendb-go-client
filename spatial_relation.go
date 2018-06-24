package ravendb

type SpatialRelation = string

const (
	SpatialRelation_WITHIN     = "Within"
	SpatialRelation_CONTAINS   = "Contains"
	SpatialRelation_DISJOINT   = "Disjoint"
	SpatialRelation_INTERSECTS = "Intersects"
)
