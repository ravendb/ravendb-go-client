package ravendb

var _ QueryToken = &NegateToken{}

var (
	NegateToken_INSTANCE = NewNegateToken()
)

type NegateToken struct {
}

func NewNegateToken() *NegateToken {
	return &NegateToken{}
}

func (t *NegateToken) writeTo(writer *StringBuilder) {
	writer.append("not")
}
