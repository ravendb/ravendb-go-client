package ravendb

var (
	DistinctToken_INSTANCE = NewDistinctToken()
)

type DistinctToken struct {
}

func NewDistinctToken() *DistinctToken {
	return &DistinctToken{}
}

func (t *DistinctToken) writeTo(writer *StringBuilder) {
	writer.append("distinct")
}
