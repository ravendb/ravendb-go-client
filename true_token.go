package ravendb

type TrueToken struct {
	*QueryToken
}

func NewTrueToken() *TrueToken {
	return &TrueToken{
		QueryToken: NewQueryToken(),
	}
}

func (t *TrueToken) writeTo(writer *StringBuilder) {
	writer.append("true")
}
