package ravendb

type PatchRequest struct {
	Script string
	Values map[string]Object
}

func (r *PatchRequest) GetScript() string {
	return r.Script
}

func (r *PatchRequest) SetScript(script string) {
	r.Script = script
}

func (r *PatchRequest) GetValues() map[string]Object {
	return r.Values
}

func (r *PatchRequest) SetValues(values map[string]Object) {
	r.Values = values
}

func NewPatchRequest() *PatchRequest {
	return &PatchRequest{}
}

func PatchRequest_forScript(script string) *PatchRequest {
	return &PatchRequest{
		Script: script,
	}
}

func (r *PatchRequest) Serialize() ObjectNode {
	values := r.Values
	if values == nil {
		values = ObjectNode{}
	}
	m := map[string]interface{}{
		"Script": r.Script,
		"Values": values,
	}
	return m
}
