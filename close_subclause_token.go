package ravendb

import "strings"

var _ queryToken = &closeSubclauseToken{}

var (
	closeSubclauseTokenInstance = &closeSubclauseToken{}
)

type closeSubclauseToken struct {
}

func (t *closeSubclauseToken) writeTo(writer *strings.Builder) {
	writer.WriteString(")")
}
