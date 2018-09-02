package ravendb

import "strings"

var _ QueryToken = &DeclareToken{}

type DeclareToken struct {
	name       string
	parameters string
	body       string
}

func NewDeclareToken(name string, body string, parameters string) *DeclareToken {
	return &DeclareToken{
		name:       name,
		body:       body,
		parameters: parameters,
	}
}

func DeclareToken_create(name string, body string) *DeclareToken {
	return DeclareToken_create2(name, body, "")
}

func DeclareToken_create2(name string, body string, parameters string) *DeclareToken {
	return NewDeclareToken(name, body, parameters)
}

func (t *DeclareToken) WriteTo(writer *strings.Builder) {

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
