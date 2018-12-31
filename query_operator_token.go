package ravendb

import "strings"

var _ queryToken = &queryOperatorToken{}

type queryOperatorToken struct {
	queryOperator QueryOperator
}

var (
	queryOperatorTokenAnd = NewQueryOperatorToken(QueryOperatorAnd)
	queryOperatorTokenOr  = NewQueryOperatorToken(QueryOperatorOr)
)

func NewQueryOperatorToken(queryOperator QueryOperator) *queryOperatorToken {
	return &queryOperatorToken{
		queryOperator: queryOperator,
	}
}

func (t *queryOperatorToken) writeTo(writer *strings.Builder) {
	if t.queryOperator == QueryOperatorAnd {
		writer.WriteString("and")
		return
	}

	writer.WriteString("or")
}
