package ravendb

import "time"

type ScriptResolver struct {
	Script           string    `json:"Script"`
	LastModifiedTime time.Time `json:"LastModifiedTime"` // TODO: our Time?
}

func (r *ScriptResolver) getScript() string {
	return r.Script
}

func (r *ScriptResolver) setScript(script string) {
	r.Script = script
}

func (r *ScriptResolver) getLastModifiedTime() time.Time {
	return r.LastModifiedTime
}

func (r *ScriptResolver) setLastModifiedTime(lastModifiedTime time.Time) {
	r.LastModifiedTime = lastModifiedTime
}
