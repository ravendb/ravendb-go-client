package ravendb

import "strings"

var _ queryToken = &suggestToken{}

type suggestToken struct {
	fieldName            string
	termParameterName    string
	optionsParameterName string
}

func (t *suggestToken) writeTo(writer *strings.Builder) {
	writer.WriteString("suggest(")
	writer.WriteString(t.fieldName)
	writer.WriteString(", $")
	writer.WriteString(t.termParameterName)

	if t.optionsParameterName != "" {
		writer.WriteString(", $")
		writer.WriteString(t.optionsParameterName)
	}

	writer.WriteString(")")
}
