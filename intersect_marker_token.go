package ravendb

var (
	IntersectMarkerToken_INSTANCE = NewIntersectMarkerToken()
)

type IntersectMarkerToken struct {
}

func NewIntersectMarkerToken() *IntersectMarkerToken {
	return &IntersectMarkerToken{}
}

func (t *IntersectMarkerToken) writeTo(writer *StringBuilder) {
	writer.append(",")
}
