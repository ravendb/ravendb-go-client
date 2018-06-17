package ravendb

import (
	"bytes"
)

// Note: StringBuilder is to make porting Java code easier

type StringBuilder struct {
	buf bytes.Buffer
}

func NewStringBuilder() *StringBuilder {
	return &StringBuilder{}
}

func (b *StringBuilder) append(s string) {
	b.buf.WriteString(s)
}
