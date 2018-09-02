package ravendb

import "strings"

var _ QueryToken = &CloseSubclauseToken{}

var (
	CloseSubclauseToken_INSTANCE = NewCloseSubclauseToken()
)

type CloseSubclauseToken struct {
}

func NewCloseSubclauseToken() *CloseSubclauseToken {
	return &CloseSubclauseToken{}
}

func (t *CloseSubclauseToken) WriteTo(writer *strings.Builder) {
	writer.WriteString(")")
}
