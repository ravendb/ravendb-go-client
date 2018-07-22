package ravendb

import (
	"bytes"
	"fmt"
	"strconv"
)

// Note: StringBuilder is to make porting Java code easier

type StringBuilder struct {
	buf bytes.Buffer
}

func NewStringBuilder() *StringBuilder {
	return &StringBuilder{}
}

func (b *StringBuilder) append(s interface{}) *StringBuilder {
	toAppend := ""
	switch v := s.(type) {
	case string:
		toAppend = v
	case int:
		toAppend = strconv.Itoa(v)
	case float64:
		toAppend = fmt.Sprintf("%s", v)
	default:
		panicIf(true, "unsupported type %T", s)
	}
	b.buf.WriteString(toAppend)
	return b
}

func (b *StringBuilder) String() string {
	return string(b.buf.Bytes())
}
