package ravendb

var (
	IntersectToken_INSTANCE = NewIntersectToken()
)

type IntersectToken struct {
}

func NewIntersectToken() *IntersectToken {
	return &IntersectToken{}
}

func (t *IntersectToken) writeTo(writer *StringBuilder) {
	writer.append(",")
}
