package ravendb

var _ QueryToken = &LoadToken{}

type LoadToken struct {
	argument string
	alias    string
}

func NewLoadToken(argument string, alias string) *LoadToken {
	return &LoadToken{
		argument: argument,
		alias:    alias,
	}
}

func LoadToken_create(argument string, alias string) *LoadToken {
	return NewLoadToken(argument, alias)
}

func (t *LoadToken) WriteTo(writer *StringBuilder) {
	writer.append(t.argument)
	writer.append(" as ")
	writer.append(t.alias)
}
