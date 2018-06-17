package ravendb

var (
	NegateToken_INSTANCE = NewNegateToken()
)

type NegateToken struct {
	*QueryToken
}

func NewNegateToken() *NegateToken {
	return &NegateToken{
		QueryToken: NewQueryToken(),
	}
}

func (t *NegateToken) writeTo(writer *StringBuilder) {
	writer.append("not")
}
