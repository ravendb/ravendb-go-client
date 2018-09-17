package ravendb

// IndexChange describes a change to the index. Can be used as DatabaseChange.
type IndexChange struct {
	Type IndexChangeTypes
	Name string
}
