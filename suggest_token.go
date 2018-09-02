package ravendb

import "strings"

var _ QueryToken = &SuggestToken{}

type SuggestToken struct {
	_fieldName            string
	_termParameterName    string
	_optionsParameterName string
}

func NewSuggestToken(fieldName string, termParameterName string, optionsParameterName string) *SuggestToken {
	return &SuggestToken{
		_fieldName:            fieldName,
		_termParameterName:    termParameterName,
		_optionsParameterName: optionsParameterName,
	}
}

func SuggestToken_create(fieldName string, termParameterName string, optionsParameterName string) *SuggestToken {
	return NewSuggestToken(fieldName, termParameterName, optionsParameterName)
}

func (t *SuggestToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("suggest(")
	writer.WriteString(t._fieldName)
	writer.WriteString(", $")
	writer.WriteString(t._termParameterName)

	if t._optionsParameterName != "" {
		writer.WriteString(", $")
		writer.WriteString(t._optionsParameterName)
	}

	writer.WriteString(")")
}
