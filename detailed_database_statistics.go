package ravendb

type DetailedDatabaseStatistics struct {
	DatabaseStatistics

	CountOfIdentities      int64 `json:"CountOfIdentities"`
	CountOfCompareExchange int64 `json:"CountOfCompareExchange"`
}
