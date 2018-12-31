package ravendb

type IndexType = string

const (
	IndexTypeNone          = "None"
	IndexTypeAutoMap       = "AutoMap"
	IndexTypeAutoMapReduce = "AutoMapReduce"
	IndexTypeMap           = "Map"
	IndexTypeMapReduce     = "MapReduce"
	IndexTypeFaulty        = "Faulty"
)
