package ravendb

func IncludesUtil_include(document ObjectNode, include string, loadId Consumer) {
	if StringUtils_isEmpty(include) || document == nil {
		return
	}

	//TBD:
}
