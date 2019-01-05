package ravendb

func includesUtilInclude(document map[string]interface{}, include string, loadID func(string)) {
	if stringIsEmpty(include) || document == nil {
		return
	}

	//TBD:
}
