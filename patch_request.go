package ravendb

type PatchRequest struct {
	script string
	values map[string]Object
}

func (r *PatchRequest) getScript() string {
	return r.script
}

func (r *PatchRequest) setScript(script string) {
	r.script = script
}

func (r *PatchRequest) getValues() map[string]Object {
	return r.values
}

func (r *PatchRequest) setValues(values map[string]Object) {
	r.values = values
}

func NewPatchRequest() *PatchRequest {
	return &PatchRequest{}
}

func PatchRequest_forScript(script string) *PatchRequest {
	return &PatchRequest{
		script: script,
	}
}

func (r *PatchRequest) serialize() ObjectNode {
	values := r.values
	if values == nil {
		values = ObjectNode{}
	}
	m := map[string]interface{}{
		"Script": r.script,
		"Values": values,
	}
	return m
}
