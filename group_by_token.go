package ravendb

import "strings"

var _ queryToken = &groupByToken{}

type groupByToken struct {
	fieldName string
	method    GroupByMethod
}

func createGroupByToken(fieldName string, method GroupByMethod) *groupByToken {
	return &groupByToken{
		fieldName: fieldName,
		method:    method,
	}
}

func (t *groupByToken) writeTo(writer *strings.Builder) {
	_method := t.method
	if _method != GroupByMethodNone {
		writer.WriteString("Array(")
	}
	writeQueryTokenField(writer, t.fieldName)
	if _method != GroupByMethodNone {
		writer.WriteString(")")
	}
}
