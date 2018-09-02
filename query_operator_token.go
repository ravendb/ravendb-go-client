package ravendb

import "strings"

var _ QueryToken = &QueryOperatorToken{}

type QueryOperatorToken struct {
	_queryOperator QueryOperator
}

var (
	QueryOperatorToken_AND = NewQueryOperatorToken(QueryOperator_AND)
	QueryOperatorToken_OR  = NewQueryOperatorToken(QueryOperator_OR)
)

func NewQueryOperatorToken(queryOperator QueryOperator) *QueryOperatorToken {
	return &QueryOperatorToken{
		_queryOperator: queryOperator,
	}
}

func (t *QueryOperatorToken) WriteTo(writer *strings.Builder) {
	if t._queryOperator == QueryOperator_AND {
		writer.WriteString("and")
		return
	}

	writer.WriteString("or")
}
