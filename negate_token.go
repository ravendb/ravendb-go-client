package ravendb

import "strings"

var _ queryToken = &negateToken{}

var (
	negateTokenInstance = &negateToken{}
)

type negateToken struct {
}

func (t *negateToken) writeTo(writer *strings.Builder) {
	writer.WriteString("not")
}
