package ravendb

import "encoding/json"

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

func (r *PatchRequest) serialize() []byte {
	m := map[string]interface{}{
		"Script": r.script,
		"Values": r.values,
	}
	d, err := json.Marshal(m)
	panicIf(err != nil, "json.Marshal() failed with %s", err)
	return d
}
