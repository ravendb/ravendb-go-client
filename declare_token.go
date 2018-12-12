package ravendb

import "strings"

var _ queryToken = &declareToken{}

type declareToken struct {
	name       string
	parameters string
	body       string
}

/*
// TODO: why is this unused? Should declareToken be publicly exposed?
func newDeclareToken(name string, body string, parameters string) *declareToken {
	return &declareToken{
		name:       name,
		body:       body,
		parameters: parameters,
	}
}
*/

func (t *declareToken) writeTo(writer *strings.Builder) {

	writer.WriteString("declare ")
	writer.WriteString("function ")
	writer.WriteString(t.name)
	writer.WriteString("(")
	writer.WriteString(t.parameters)
	writer.WriteString(") ")
	writer.WriteString("{")
	writer.WriteString("\n")
	writer.WriteString(t.body)
	writer.WriteString("\n")
	writer.WriteString("}")
	writer.WriteString("\n")
}
