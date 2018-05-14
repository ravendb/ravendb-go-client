package ravendb

// tryGetIDFromInstance returns value of ID field on struct if it's of type
// string. Returns empty string if there's no ID field or it's not string
func tryGetIDFromInstance(entity interface{}) string {
	// TODO: implement me
	panicIf(true, "NYI")
	return ""
}

// trySetIDOnEnity tries to set value of ID field on struct to id
// returns false if entity has no ID field or if it's not string
func trySetIDOnEntity(entity interface{}, id string) bool {
	// TODO: implement me
	panicIf(true, "NYI")
	return false
}
