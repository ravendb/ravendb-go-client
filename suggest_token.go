package ravendb

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

func (t *SuggestToken) WriteTo(writer *StringBuilder) {
	writer.append("suggest(")
	writer.append(t._fieldName)
	writer.append(", $")
	writer.append(t._termParameterName)

	if t._optionsParameterName != "" {
		writer.append(", $")
		writer.append(t._optionsParameterName)
	}

	writer.append(")")
}
