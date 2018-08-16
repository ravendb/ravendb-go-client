package ravendb

// TODO: desugar
type IndexConfiguration = map[string]string

func NewIndexConfiguration() IndexConfiguration {
	return make(map[string]string)
}
