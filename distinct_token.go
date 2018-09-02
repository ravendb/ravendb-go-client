package ravendb

import "strings"

var _ QueryToken = &DistinctToken{}

var (
	DistinctToken_INSTANCE = NewDistinctToken()
)

type DistinctToken struct {
}

func NewDistinctToken() *DistinctToken {
	return &DistinctToken{}
}

func (t *DistinctToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("distinct")
}
