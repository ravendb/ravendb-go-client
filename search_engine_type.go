package ravendb

type SearchEngineType = string

const (
	SearchEngineTypeNone   = "None"
	SearchEngineTypeLucene = "Lucene"
	SearchEngineTypeCorax  = "Corax"
)
