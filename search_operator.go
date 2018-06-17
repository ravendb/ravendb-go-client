package ravendb

type SearchOperator int

const (
	SearchOperator_OR SearchOperator = iota
	SearchOperator_AND
)
