package ravendb

import "strings"

var _ queryToken = &distinctToken{}

var (
	distinctTokenInstance = &distinctToken{}
)

type distinctToken struct {
}

func (t *distinctToken) writeTo(writer *strings.Builder) {
	writer.WriteString("distinct")
}
