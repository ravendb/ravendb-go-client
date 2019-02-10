package ravendb

import "strings"

var _ queryToken = &openSubclauseToken{}

var (
	openSubclauseTokenInstance = &openSubclauseToken{}
)

type openSubclauseToken struct {
}

func (t *openSubclauseToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("(")
	return nil
}
