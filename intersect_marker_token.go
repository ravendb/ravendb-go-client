package ravendb

import "strings"

var _ queryToken = &intersectMarkerToken{}

var (
	intersectMarkerTokenInstance = &intersectMarkerToken{}
)

type intersectMarkerToken struct {
}

func (t *intersectMarkerToken) writeTo(writer *strings.Builder) {
	writer.WriteString(",")
}
