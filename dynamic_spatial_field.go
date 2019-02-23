package ravendb

// DynamicSpatialField is an interface for implementing name of the field to
// be queried
type DynamicSpatialField interface {
	// ToField returns a name of the field used in queries
	ToField(ensureValidFieldName func(string, bool) string) string
}
