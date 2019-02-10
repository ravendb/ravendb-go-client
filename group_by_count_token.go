package ravendb

import "strings"

var _ queryToken = &groupByCountToken{}

type groupByCountToken struct {
	fieldName string
}

func (t *groupByCountToken) writeTo(writer *strings.Builder) error {

	writer.WriteString("count()")

	if t.fieldName == "" {
		return nil
	}

	writer.WriteString(" as ")
	writer.WriteString(t.fieldName)
	return nil
}
