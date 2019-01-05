package ravendb

import "strings"

var _ queryToken = &groupBySumToken{}

type groupBySumToken struct {
	projectedName string
	fieldName     string
}

func newGroupBySumToken(fieldName string, projectedName string) *groupBySumToken {
	return &groupBySumToken{
		fieldName:     fieldName,
		projectedName: projectedName,
	}
}

func createGroupBySumToken(fieldName string, projectedName string) *groupBySumToken {
	return newGroupBySumToken(fieldName, projectedName)
}

func (t *groupBySumToken) writeTo(writer *strings.Builder) {
	writer.WriteString("sum(")
	writer.WriteString(t.fieldName)
	writer.WriteString(")")

	if t.projectedName == "" {
		return
	}

	writer.WriteString(" as ")
	writer.WriteString(t.projectedName)
}
