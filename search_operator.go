package ravendb

type SearchOperator int

const (
	SearchOperatorUnset SearchOperator = iota
	SearchOperatorOr
	SearchOperatorAnd
)
