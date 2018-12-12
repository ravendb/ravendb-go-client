package ravendb

type GeographySpatialOptionsFactory struct {
	// yes, empty
}

func NewGeographySpatialOptionsFactory() *GeographySpatialOptionsFactory {
	return &GeographySpatialOptionsFactory{}
}

func (f *GeographySpatialOptionsFactory) defaultOptions() *SpatialOptions {
	return f.defaultOptionsWithRadius(SpatialUnits_KILOMETERS)
}

func (f *GeographySpatialOptionsFactory) defaultOptionsWithRadius(circleRadiusUnits SpatialUnits) *SpatialOptions {
	return f.geohashPrefixTreeIndexWithRadius(0, circleRadiusUnits)
}

func (f *GeographySpatialOptionsFactory) boundingBoxIndex() *SpatialOptions {
	return f.boundingBoxIndexWithRadius(SpatialUnits_KILOMETERS)
}

func (f *GeographySpatialOptionsFactory) boundingBoxIndexWithRadius(circleRadiusUnits SpatialUnits) *SpatialOptions {
	ops := NewSpatialOptions()
	ops.Type = SpatialFieldType_GEOGRAPHY
	ops.Strategy = SpatialSearchStrategy_BOUNDING_BOX
	ops.Units = circleRadiusUnits
	return ops
}

func (f *GeographySpatialOptionsFactory) geohashPrefixTreeIndex(maxTreeLevel int) *SpatialOptions {
	return f.geohashPrefixTreeIndexWithRadius(maxTreeLevel, SpatialUnits_KILOMETERS)
}

func (f *GeographySpatialOptionsFactory) geohashPrefixTreeIndexWithRadius(maxTreeLevel int, circleRadiusUnits SpatialUnits) *SpatialOptions {
	if maxTreeLevel == 0 {
		maxTreeLevel = SpatialOptions_DEFAULT_GEOHASH_LEVEL
	}

	opts := NewSpatialOptions()
	opts.Type = SpatialFieldType_GEOGRAPHY
	opts.MaxTreeLevel = maxTreeLevel
	opts.Strategy = SpatialSearchStrategy_GEOHASH_PREFIX_TREE
	opts.Units = circleRadiusUnits
	return opts
}

func (f *GeographySpatialOptionsFactory) quadPrefixTreeIndex(maxTreeLevel int) *SpatialOptions {
	return f.quadPrefixTreeIndexWithRadius(maxTreeLevel, SpatialUnits_KILOMETERS)
}

func (f *GeographySpatialOptionsFactory) quadPrefixTreeIndexWithRadius(maxTreeLevel int, circleRadiusUnits SpatialUnits) *SpatialOptions {
	if maxTreeLevel == 0 {
		maxTreeLevel = SpatialOptions_DEFAULT_QUAD_TREE_LEVEL
	}

	opts := NewSpatialOptions()
	opts.Type = SpatialFieldType_GEOGRAPHY
	opts.MaxTreeLevel = maxTreeLevel
	opts.Strategy = SpatialSearchStrategy_QUAD_PREFIX_TREE
	opts.Units = circleRadiusUnits
	return opts
}

type CartesianSpatialOptionsFactory struct {
	// yes, empty
}

func NewCartesianSpatialOptionsFactory() *CartesianSpatialOptionsFactory {
	return &CartesianSpatialOptionsFactory{}
}

func (f *CartesianSpatialOptionsFactory) boundingBoxIndex() *SpatialOptions {
	opts := NewSpatialOptions()
	opts.Type = SpatialFieldType_CARTESIAN
	opts.Strategy = SpatialSearchStrategy_BOUNDING_BOX
	return opts
}

func (f *CartesianSpatialOptionsFactory) quadPrefixTreeIndex(maxTreeLevel int, bounds *SpatialBounds) *SpatialOptions {
	panicIf(maxTreeLevel == 0, "maxTreeLevel cannot be 0")

	opts := NewSpatialOptions()
	opts.Type = SpatialFieldType_CARTESIAN
	opts.MaxTreeLevel = maxTreeLevel
	opts.Strategy = SpatialSearchStrategy_QUAD_PREFIX_TREE
	opts.MinX = bounds.minX
	opts.MinY = bounds.minY
	opts.MaxX = bounds.maxX
	opts.MaxY = bounds.maxY

	return opts
}

type SpatialBounds struct {
	minX float64
	maxX float64
	minY float64
	maxY float64
}

func NewSpatialBounds(minX float64, minY float64, maxX float64, maxY float64) *SpatialBounds {
	return &SpatialBounds{
		minX: minX,
		maxX: maxX,
		minY: minY,
		maxY: maxY,
	}
}

type SpatialOptionsFactory struct {
	// yes, empty
}

func NewSpatialOptionsFactory() *SpatialOptionsFactory {
	return &SpatialOptionsFactory{}
}

func (f *SpatialOptionsFactory) getGeography() *GeographySpatialOptionsFactory {
	return NewGeographySpatialOptionsFactory()
}

func (f *SpatialOptionsFactory) getCartesian() *CartesianSpatialOptionsFactory {
	return NewCartesianSpatialOptionsFactory()
}
