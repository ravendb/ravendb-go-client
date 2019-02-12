package ravendb

import "strings"

var _ queryToken = &queryOperatorToken{}

type queryOperatorToken struct {
	queryOperator QueryOperator
}

var (
	queryOperatorTokenAnd = &queryOperatorToken{
		queryOperator: queryOperatorAnd,
	}
	queryOperatorTokenOr = &queryOperatorToken{
		queryOperator: queryOperatorOr,
	}
)

func (t *queryOperatorToken) writeTo(writer *strings.Builder) error {
	if t.queryOperator == queryOperatorAnd {
		writer.WriteString("and")
		return nil
	}

	writer.WriteString("or")
	return nil
}
