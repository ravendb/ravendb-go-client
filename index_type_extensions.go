package ravendb

func IndexTypeExtensions_isMap(typ IndexType) bool {
	switch typ {
	case IndexTypeMap, IndexTypeAutoMap, IndexTypeJavaScriptMap:
		return true
	}
	return false
}

func IndexTypeExtensions_isMapReduce(typ IndexType) bool {
	switch typ {
	case IndexTypeMapReduce, IndexTypeAutoMapReduce, IndexTypeJavaScriptMapReduce:
		return true
	}
	return false
}

func IndexTypeExtensions_isAuto(typ IndexType) bool {
	switch typ {
	case IndexTypeAutoMap, IndexTypeAutoMapReduce:
		return true
	}
	return false
}

func IndexTypeExtensions_isState(typ IndexType) bool {
	switch typ {
	case IndexTypeMap, IndexTypeMapReduce,
		IndexTypeJavaScriptMap, IndexTypeJavaScriptMapReduce:
		return true
	}
	return false
}

func IndexTypeExtensions_isJavaScript(typ IndexType) bool {
	switch typ {
	case IndexTypeJavaScriptMap, IndexTypeJavaScriptMapReduce:
		return true
	}
	return false
}
