package ravendb

type DeclareToken struct {
	*QueryToken

	name       string
	parameters string
	body       string
}

func NewDeclareToken(name string, body string, parameters string) *DeclareToken {
	return &DeclareToken{
		QueryToken: NewQueryToken(),
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

func (t *DeclareToken) writeTo(writer *StringBuilder) {

	writer.append("declare ")
	writer.append("function ")
	writer.append(t.name)
	writer.append("(")
	writer.append(t.parameters)
	writer.append(") ")
	writer.append("{")
	writer.append("\n")
	writer.append(t.body)
	writer.append("\n")
	writer.append("}")
	writer.append("\n")
}
