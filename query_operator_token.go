package ravendb

import "strings"

var _ queryToken = &queryOperatorToken{}

type queryOperatorToken struct {
	queryOperator QueryOperator
}

var (
	queryOperatorTokenAnd = &queryOperatorToken{
		queryOperator: QueryOperatorAnd,
	}
	queryOperatorTokenOr = &queryOperatorToken{
		queryOperator: QueryOperatorOr,
	}
)

func (t *queryOperatorToken) writeTo(writer *strings.Builder) error {
	if t.queryOperator == QueryOperatorAnd {
		writer.WriteString("and")
		return nil
	}

	writer.WriteString("or")
	return nil
}
