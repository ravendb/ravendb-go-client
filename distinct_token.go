package ravendb

var _ QueryToken = &DistinctToken{}

var (
	DistinctToken_INSTANCE = NewDistinctToken()
)

type DistinctToken struct {
}

func NewDistinctToken() *DistinctToken {
	return &DistinctToken{}
}

func (t *DistinctToken) WriteTo(writer *StringBuilder) {
	writer.append("distinct")
}
