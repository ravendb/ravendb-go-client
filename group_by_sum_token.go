package ravendb

import "strings"

var _ QueryToken = &GroupBySumToken{}

type GroupBySumToken struct {
	_projectedName string
	_fieldName     string
}

func NewGroupBySumToken(fieldName string, projectedName string) *GroupBySumToken {
	return &GroupBySumToken{
		_fieldName:     fieldName,
		_projectedName: projectedName,
	}
}

func GroupBySumToken_create(fieldName string, projectedName string) *GroupBySumToken {
	return NewGroupBySumToken(fieldName, projectedName)
}

func (t *GroupBySumToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("sum(")
	writer.WriteString(t._fieldName)
	writer.WriteString(")")

	if t._projectedName == "" {
		return
	}

	writer.WriteString(" as ")
	writer.WriteString(t._projectedName)
}
