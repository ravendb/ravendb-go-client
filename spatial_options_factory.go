package ravendb

// NewGeographyDefaultOptions returns default geography SpatialOptions
func NewGeographyDefaultOptions() *SpatialOptions {
	return NewGeographyDefaultOptionsWithRadius(SpatialUnitsKilometers)
}

// NewGeographyDefaultOptionsWithRadius returns default geography SpatialOptions
// with a a given radius
func NewGeographyDefaultOptionsWithRadius(circleRadiusUnits SpatialUnits) *SpatialOptions {
	return NewGeographyGeohashPrefixTreeIndexWithRadius(0, circleRadiusUnits)
}

// NewGeograpyboundingBoxIndex returns geography SpatialOptions for a bounding box
func NewGeograpyboundingBoxIndex() *SpatialOptions {
	return NewGeographyBoundingBoxIndexWithRadius(SpatialUnitsKilometers)
}

// NewGeographyBoundingBoxIndexWithRadius return geography SpatialOptions
// for a bounding box with a given radius
func NewGeographyBoundingBoxIndexWithRadius(circleRadiusUnits SpatialUnits) *SpatialOptions {
	opts := NewSpatialOptions()
	opts.Type = SpatialFieldGeography
	opts.Strategy = SpatialSearchStrategyBoundingBox
	opts.Units = circleRadiusUnits
	return opts
}

// NewGeographyGeohashPrefixTreeIndex returns geography SpatialOptions for
// using geography geohash prefix tree index
func NewGeographyGeohashPrefixTreeIndex(maxTreeLevel int) *SpatialOptions {
	return NewGeographyGeohashPrefixTreeIndexWithRadius(maxTreeLevel, SpatialUnitsKilometers)
}

// NewGeographyGeohashPrefixTreeIndexWithRadius returns geography SpatialOptions for
//// using geography geohash prefix tree index with a given circle radius
func NewGeographyGeohashPrefixTreeIndexWithRadius(maxTreeLevel int, circleRadiusUnits SpatialUnits) *SpatialOptions {
	if maxTreeLevel == 0 {
		maxTreeLevel = SpatialOptionsDefaultGeohashLevel
	}

	opts := NewSpatialOptions()
	opts.Type = SpatialFieldGeography
	opts.MaxTreeLevel = maxTreeLevel
	opts.Strategy = SpatialSearchStrategyGeohashPrefixTree
	opts.Units = circleRadiusUnits
	return opts
}

// NewGeographyQuadPrefixTreeIndex returns geography SpatialOptions
// for quad prefix tree
func NewGeographyQuadPrefixTreeIndex(maxTreeLevel int) *SpatialOptions {
	return NewGeographyQuadPrefixTreeIndexWithRadius(maxTreeLevel, SpatialUnitsKilometers)
}

// NewGeographyQuadPrefixTreeIndex returns geography SpatialOptions
// for quad prefix tree with a given radius
func NewGeographyQuadPrefixTreeIndexWithRadius(maxTreeLevel int, circleRadiusUnits SpatialUnits) *SpatialOptions {
	if maxTreeLevel == 0 {
		maxTreeLevel = SpatialOptionsDefaultQuadTreeLevel
	}

	opts := NewSpatialOptions()
	opts.Type = SpatialFieldGeography
	opts.MaxTreeLevel = maxTreeLevel
	opts.Strategy = SpatialSearchStrategyQuadPrefixTree
	opts.Units = circleRadiusUnits
	return opts
}

// NewCartesianBoundingBoxIndex returns cartesian SpatialOptions
func NewCartesianBoundingBoxIndex() *SpatialOptions {
	opts := NewSpatialOptions()
	opts.Type = SpatialFieldCartesian
	opts.Strategy = SpatialSearchStrategyBoundingBox
	return opts
}

// NewCartesianQuadPrefixTreeIndex returns cartesian SpatialOptions for
// quad prefix tree index
func NewCartesianQuadPrefixTreeIndex(maxTreeLevel int, bounds *SpatialBounds) *SpatialOptions {
	panicIf(maxTreeLevel > 0, "maxTreeLevel must be > 0")
	opts := NewSpatialOptions()
	opts.Type = SpatialFieldCartesian
	opts.MaxTreeLevel = maxTreeLevel
	opts.Strategy = SpatialSearchStrategyQuadPrefixTree
	opts.MinX = bounds.minX
	opts.MinY = bounds.minY
	opts.MaxX = bounds.maxX
	opts.MaxY = bounds.maxY
	return opts
}

// SpatialBounds describes bounds of a region
type SpatialBounds struct {
	minX float64
	maxX float64
	minY float64
	maxY float64
}
