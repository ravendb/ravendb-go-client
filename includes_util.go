package ravendb

func IncludesUtil_include(document ObjectNode, include string, loadId func(string)) {
	if stringIsEmpty(include) || document == nil {
		return
	}

	//TBD:
}
