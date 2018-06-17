package ravendb

var (
	DistinctToken_INSTANCE = NewDistinctToken()
)

type DistinctToken struct {
	*QueryToken
}

func NewDistinctToken() *DistinctToken {
	return &DistinctToken{
		QueryToken: NewQueryToken(),
	}
}

func (t *DistinctToken) writeTo(writer *StringBuilder) {
	writer.append("distinct")
}
