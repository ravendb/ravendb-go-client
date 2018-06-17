package ravendb

type TrueToken struct {
}

func NewTrueToken() *TrueToken {
	return &TrueToken{}
}

func (t *TrueToken) writeTo(writer *StringBuilder) {
	writer.append("true")
}
