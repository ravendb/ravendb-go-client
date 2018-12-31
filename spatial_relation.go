package ravendb

type SpatialRelation = string

const (
	SpatialRelationWithin     = "Within"
	SpatialRelationContains   = "Contains"
	SpatialRelationDisjoin    = "Disjoint"
	SpatialRelationIntersects = "Intersects"
)
