package ravendb

func IncludesUtil_include(document ObjectNode, include string, loadId func(string)) {
	if StringUtils_isEmpty(include) || document == nil {
		return
	}

	//TBD:
}
