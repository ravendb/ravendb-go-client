package ravendb

type DynamicSpatialField interface {
	toField(ensureValidFieldName func(string, bool) string) string
}
