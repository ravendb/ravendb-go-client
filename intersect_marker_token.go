package ravendb

var _ QueryToken = &IntersectMarkerToken{}

var (
	IntersectMarkerToken_INSTANCE = NewIntersectMarkerToken()
)

type IntersectMarkerToken struct {
}

func NewIntersectMarkerToken() *IntersectMarkerToken {
	return &IntersectMarkerToken{}
}

func (t *IntersectMarkerToken) WriteTo(writer *StringBuilder) {
	writer.append(",")
}
