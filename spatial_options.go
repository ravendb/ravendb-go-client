package ravendb

const (
	//about 4.78 meters at equator, should be good enough (see: http://unterbahn.com/2009/11/metric-dimensions-of-geohash-partitions-at-the-equator/)
	SpatialOptionsDefaultGeohashLevel = 9
	//about 4.78 meters at equator, should be good enough
	SpatialOptionsDefaultQuadTreeLevel = 23
)

/// SpatialOptions describes spatial options
type SpatialOptions struct {
	Type         SpatialFieldType      `json:"Type"`
	Strategy     SpatialSearchStrategy `json:"Strategy"`
	MaxTreeLevel int                   `json:"MaxTreeLevel"`
	MinX         float64               `json:"MinX"`
	MaxX         float64               `json:"MaxX"`
	MinY         float64               `json:"MinY"`
	MaxY         float64               `json:"MaxY"`

	// Circle radius units, only used for geography  indexes
	Units SpatialUnits `json:"Units"`
}

// NewSpatialOptions returns new SpatialOptions with default values
func NewSpatialOptions() *SpatialOptions {
	return &SpatialOptions{
		Type:         SpatialFieldGeography,
		Strategy:     SpatialSearchStrategyGeohashPrefixTree,
		MaxTreeLevel: SpatialOptionsDefaultGeohashLevel,
		MinX:         -180,
		MaxX:         180,
		MinY:         -90,
		MaxY:         90,
		Units:        SpatialUnitsKilometers,
	}
}
