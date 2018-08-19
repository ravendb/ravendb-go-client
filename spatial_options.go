package ravendb

const (
	//about 4.78 meters at equator, should be good enough (see: http://unterbahn.com/2009/11/metric-dimensions-of-geohash-partitions-at-the-equator/)
	SpatialOptions_DEFAULT_GEOHASH_LEVEL = 9
	//about 4.78 meters at equator, should be good enough
	SpatialOptions_DEFAULT_QUAD_TREE_LEVEL = 23
)

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

func NewSpatialOptions() *SpatialOptions {
	return &SpatialOptions{
		Type:         SpatialFieldType_GEOGRAPHY,
		Strategy:     SpatialSearchStrategy_GEOHASH_PREFIX_TREE,
		MaxTreeLevel: SpatialOptions_DEFAULT_GEOHASH_LEVEL,
		MinX:         -180,
		MaxX:         180,
		MinY:         -90,
		MaxY:         90,
		Units:        SpatialUnits_KILOMETERS,
	}
}

func SpatialOptionsDup(options *SpatialOptions) *SpatialOptions {
	var res SpatialOptions = *options
	return &res
}
