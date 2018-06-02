package ravendb

import "time"

type ScriptResolver struct {
	script           string    `json:"Script"`
	lastModifiedTime time.Time `json:"LastModifiedTime"` // TODO: ServerTime?
}

func (r *ScriptResolver) getScript() String {
	return r.script
}

func (r *ScriptResolver) setScript(script String) {
	r.script = script
}

func (r *ScriptResolver) getLastModifiedTime() time.Time {
	return r.lastModifiedTime
}

func (r *ScriptResolver) setLastModifiedTime(lastModifiedTime time.Time) {
	r.lastModifiedTime = lastModifiedTime
}
