package ravendb

// SpatialSearchStrategy represents spatial search strategy
type SpatialSearchStrategy = string

const (
	SpatialSearchStrategyGeohashPrefixTree = "GeohashPrefixTree"
	SpatialSearchStrategyQuadPrefixTree    = "QuadPrefixTree"
	SpatialSearchStrategyBoundingBox       = "BoundingBox"
)
