package ravendb

type IndexType = string

const (
	IndexType_NONE            = "None"
	IndexType_AUTO_MAP        = "AutoMap"
	IndexType_AUTO_MAP_REDUCE = "AutoMapReduce"
	IndexType_MAP             = "Map"
	IndexType_MAP_REDUCE      = "MapReduce"
	IndexType_FAULTY          = "Faulty"
)
