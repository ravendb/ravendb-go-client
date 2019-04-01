package ravendb

import "strings"

var _ queryToken = &timingsToken{}

var (
	timingsToken_INSTANCE = &timingsToken{}
)

type timingsToken struct {
}

func (t *timingsToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("timings()")
	return nil
}
