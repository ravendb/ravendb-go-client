package ravendb

type QueryHighlightings struct {
	_highlightings []*Highlightings
}

func (h *QueryHighlightings) add(fieldName string) *Highlightings {
	fieldHighlightings := NewHighlightings(fieldName)
	h._highlightings = append(h._highlightings, fieldHighlightings)
	return fieldHighlightings
}

func (h *QueryHighlightings) update(queryResult *QueryResult) {
	for _, fieldHighlightings := range h._highlightings {
		fieldHighlightings.update(queryResult.Highlightings)
	}
}
