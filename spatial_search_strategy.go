package ravendb

type SpatialSearchStrategy = string

const (
	SpatialSearchStrategy_GEOHASH_PREFIX_TREE = "GeohashPrefixTree"
	SpatialSearchStrategy_QUAD_PREFIX_TREE    = "QuadPrefixTree"
	SpatialSearchStrategy_BOUNDING_BOX        = "BoundingBox"
)
