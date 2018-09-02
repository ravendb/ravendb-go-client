package ravendb

import "strings"

var _ QueryToken = &NegateToken{}

var (
	NegateToken_INSTANCE = NewNegateToken()
)

type NegateToken struct {
}

func NewNegateToken() *NegateToken {
	return &NegateToken{}
}

func (t *NegateToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("not")
}
