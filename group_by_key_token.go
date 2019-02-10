package ravendb

import "strings"

var _ queryToken = &groupByKeyToken{}

type groupByKeyToken struct {
	fieldName     string
	projectedName string
}

func newGroupByKeyToken(fieldName string, projectedName string) *groupByKeyToken {
	return &groupByKeyToken{
		fieldName:     fieldName,
		projectedName: projectedName,
	}
}

func createGroupByKeyToken(fieldName string, projectedName string) *groupByKeyToken {
	return newGroupByKeyToken(fieldName, projectedName)
}

func (t *groupByKeyToken) writeTo(writer *strings.Builder) error {
	writeQueryTokenField(writer, firstNonEmptyString(t.fieldName, "key()"))

	if t.projectedName == "" || t.projectedName == t.fieldName {
		return nil
	}

	writer.WriteString(" as ")
	writer.WriteString(t.projectedName)

	return nil
}
