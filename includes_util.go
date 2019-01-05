package ravendb

func IncludesUtil_include(document map[string]interface{}, include string, loadID func(string)) {
	if stringIsEmpty(include) || document == nil {
		return
	}

	//TBD:
}
