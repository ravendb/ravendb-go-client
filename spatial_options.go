package ravendb

const (
	//about 4.78 meters at equator, should be good enough (see: http://unterbahn.com/2009/11/metric-dimensions-of-geohash-partitions-at-the-equator/)
	SpatialOptions_DEFAULT_GEOHASH_LEVEL = 9
	//about 4.78 meters at equator, should be good enough
	SpatialOptions_DEFAULT_QUAD_TREE_LEVEL = 23
)

type SpatialOptions struct {
	typ          SpatialFieldType
	strategy     SpatialSearchStrategy
	maxTreeLevel int
	minX         float64
	maxX         float64
	minY         float64
	maxY         float64

	// Circle radius units, only used for geography  indexes
	units SpatialUnits
}

func NewSpatialOptions() *SpatialOptions {
	return &SpatialOptions{
		typ:          SpatialFieldType_GEOGRAPHY,
		strategy:     SpatialSearchStrategy_GEOHASH_PREFIX_TREE,
		maxTreeLevel: SpatialOptions_DEFAULT_GEOHASH_LEVEL,
		minX:         -180,
		maxX:         180,
		minY:         -90,
		maxY:         90,
		units:        SpatialUnits_KILOMETERS,
	}
}

func SpatialOptionsDup(options *SpatialOptions) *SpatialOptions {
	var res SpatialOptions = *options
	return &res
}

func (o *SpatialOptions) GetType() SpatialFieldType {
	return o.typ
}

func (o *SpatialOptions) SetType(typ SpatialFieldType) {
	o.typ = typ
}

func (o *SpatialOptions) GetStrategy() SpatialSearchStrategy {
	return o.strategy
}

func (o *SpatialOptions) SetStrategy(strategy SpatialSearchStrategy) {
	o.strategy = strategy
}

func (o *SpatialOptions) GetMaxTreeLevel() int {
	return o.maxTreeLevel
}

func (o *SpatialOptions) SetMaxTreeLevel(maxTreeLevel int) {
	o.maxTreeLevel = maxTreeLevel
}

func (o *SpatialOptions) GetMinX() float64 {
	return o.minX
}

func (o *SpatialOptions) SetMinX(minX float64) {
	o.minX = minX
}

func (o *SpatialOptions) GetMaxX() float64 {
	return o.maxX
}

func (o *SpatialOptions) SetMaxX(maxX float64) {
	o.maxX = maxX
}

func (o *SpatialOptions) GetMinY() float64 {
	return o.minY
}

func (o *SpatialOptions) SetMinY(minY float64) {
	o.minY = minY
}

func (o *SpatialOptions) GetMaxY() float64 {
	return o.maxY
}

func (o *SpatialOptions) SetMaxY(maxY float64) {
	o.maxY = maxY
}

func (o *SpatialOptions) GetUnits() SpatialUnits {
	return o.units
}

func (o *SpatialOptions) SetUnits(units SpatialUnits) {
	o.units = units
}
