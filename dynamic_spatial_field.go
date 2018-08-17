package ravendb

type DynamicSpatialField interface {
	ToField(ensureValidFieldName func(string, bool) string) string
}
