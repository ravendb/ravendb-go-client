package ravendb

import "strings"

var _ queryToken = &GroupByToken{}

type GroupByToken struct {
	_fieldName string
	_method    GroupByMethod
}

func NewGroupByToken(fieldName string, method GroupByMethod) *GroupByToken {
	return &GroupByToken{
		_fieldName: fieldName,
		_method:    method,
	}
}

func GroupByToken_create(fieldName string) *GroupByToken {
	return GroupByToken_createWithMethod(fieldName, GroupByMethod_NONE)
}

func GroupByToken_createWithMethod(fieldName string, method GroupByMethod) *GroupByToken {
	return NewGroupByToken(fieldName, method)
}

func (t *GroupByToken) writeTo(writer *strings.Builder) {
	_method := t._method
	if _method != GroupByMethod_NONE {
		writer.WriteString("Array(")
	}
	writeQueryTokenField(writer, t._fieldName)
	if _method != GroupByMethod_NONE {
		writer.WriteString(")")
	}
}
