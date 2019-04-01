package ravendb

import "strings"

var _ queryToken = &explanationToken{}

type explanationToken struct {
	_optionsParameterName string
}

func (t *explanationToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("explanations(")
	if t._optionsParameterName != "" {
		writer.WriteString("$")
		writer.WriteString(t._optionsParameterName)
	}
	writer.WriteString(")")
	return nil
}
