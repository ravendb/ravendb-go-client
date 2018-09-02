package ravendb

import "strings"

var _ queryToken = &groupByCountToken{}

type groupByCountToken struct {
	fieldName string
}

func (t *groupByCountToken) writeTo(writer *strings.Builder) {
	writer.WriteString("count()")

	if t.fieldName == "" {
		return
	}

	writer.WriteString(" as ")
	writer.WriteString(t.fieldName)
}
