package ravendb

import "strings"

var _ QueryToken = &GroupByCountToken{}

type GroupByCountToken struct {
	_fieldName string
}

func NewGroupByCountToken(fieldName string) *GroupByCountToken {
	return &GroupByCountToken{
		_fieldName: fieldName,
	}
}

func GroupByCountToken_create(fieldName string) *GroupByCountToken {
	return NewGroupByCountToken(fieldName)
}

func (t *GroupByCountToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("count()")

	if t._fieldName == "" {
		return
	}

	writer.WriteString(" as ")
	writer.WriteString(t._fieldName)
}
