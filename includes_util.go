package ravendb

func IncludesUtil_include(document ObjectNode, include string, loadID func(string)) {
	if stringIsEmpty(include) || document == nil {
		return
	}

	//TBD:
}
