package ravendb

type Highlightings struct {
	// TODO: key is case-insensitive
	_highlightings map[string][]string
	_fieldName     string
}

func NewHighlightings(fieldName string) *Highlightings {
	return &Highlightings{
		_fieldName:     fieldName,
		_highlightings: map[string][]string{},
	}
}

/*
   public Set<String> getResultIndents() {
       return _highlightings.keySet();
   }
*/

/*
   public String[] getFragments(String key) {
       String[] result = _highlightings.get(key);
       if (result == null) {
           return new String[0];
       }
       return result;
   }
*/

func (h *Highlightings) update(highlightings map[string]map[string][]string) {
	h._highlightings = map[string][]string{}

	if highlightings == nil {
		return
	}
	result, ok := highlightings[h._fieldName]
	if !ok {
		return
	}

	for k, v := range result {
		// TODO: case-insensitive
		h._highlightings[k] = v
	}
}
