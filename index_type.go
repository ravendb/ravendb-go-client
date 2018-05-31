package ravendb

type IndexType = string

const (
	IndexType_NONE            = "NONE"
	IndexType_AUTO_MAP        = "AUTO_MAP"
	IndexType_AUTO_MAP_REDUCE = "AUTO_MAP_REDUCE"
	IndexType_MAP             = "MAP"
	IndexType_MAP_REDUCE      = "MAP_REDUCE"
	IndexType_FAULTY          = "FAULTY"
)
