package ravendb

type SearchOperator int

const (
	SearchOperator_UNSET SearchOperator = iota
	SearchOperator_OR
	SearchOperator_AND
)
