package ravendb

import "strings"

var _ QueryToken = &OpenSubclauseToken{}

var (
	OpenSubclauseToken_INSTANCE = NewOpenSubclauseToken()
)

type OpenSubclauseToken struct {
}

func NewOpenSubclauseToken() *OpenSubclauseToken {
	return &OpenSubclauseToken{}
}

func (t *OpenSubclauseToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("(")
}
