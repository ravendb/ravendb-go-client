package data

import "strings"

type PatchRequest struct {
	script string
	values []string
}
func NewPatchRequest(script string, values []string) *PatchRequest {
	return &PatchRequest{script: script, values: values }
}
func (ref *PatchRequest) ToJson() string {
	return `{"Script":` + ref.script + `, "Values":[` + strings.Join(ref.values, ",") + `]}`
}

