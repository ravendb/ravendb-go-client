package ravendb

import "strings"

var _ queryToken = &queryOperatorToken{}

type queryOperatorToken struct {
	queryOperator QueryOperator
}

var (
	QueryOperatorToken_AND = NewQueryOperatorToken(QueryOperator_AND)
	QueryOperatorToken_OR  = NewQueryOperatorToken(QueryOperator_OR)
)

func NewQueryOperatorToken(queryOperator QueryOperator) *queryOperatorToken {
	return &queryOperatorToken{
		queryOperator: queryOperator,
	}
}

func (t *queryOperatorToken) writeTo(writer *strings.Builder) {
	if t.queryOperator == QueryOperator_AND {
		writer.WriteString("and")
		return
	}

	writer.WriteString("or")
}
