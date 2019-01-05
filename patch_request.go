package ravendb

// PatchRequest represents patch request
type PatchRequest struct {
	Script string
	Values map[string]interface{}
}

// Serialize serializes PatchRequest to json
func (r *PatchRequest) Serialize() map[string]interface{} {
	values := r.Values
	if values == nil {
		values = map[string]interface{}{}
	}
	m := map[string]interface{}{
		"Script": r.Script,
		"Values": values,
	}
	return m
}
