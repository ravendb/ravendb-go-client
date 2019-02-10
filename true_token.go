package ravendb

import "strings"

var _ queryToken = &trueToken{}

var trueTokenInstance = &trueToken{}

type trueToken struct {
}

func (t *trueToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("true")
	return nil
}
