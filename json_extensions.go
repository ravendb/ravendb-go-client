package ravendb

// TODO: change the name to better reflect what it does
func JsonExtensions_writeIndexQuery(conventions *DocumentConventions, query *IndexQuery) map[string]interface{} {
	res := map[string]interface{}{}
	res["Query"] = query.GetQuery()
	if query.IsPageSizeSet() && query.GetPageSize() > 0 {
		res["PageSize"] = query.GetPageSize()
	}

	if query.IsWaitForNonStaleResults() {
		res["WaitForNonStaleResults"] = query.IsWaitForNonStaleResults()
	}

	if query.GetStart() > 0 {
		res["Start"] = query.GetStart()
	}

	if query.GetWaitForNonStaleResultsTimeout() != 0 {
		s := TimeUtils_durationToTimeSpan(query.GetWaitForNonStaleResultsTimeout())
		res["WaitForNonStaleResultsTimeout"] = s
	}

	if query.IsDisableCaching() {
		res["DisableCaching"] = query.IsDisableCaching()
	}

	if query.IsSkipDuplicateChecking() {
		res["SkipDuplicateChecking"] = query.IsSkipDuplicateChecking()
	}
	params := query.GetQueryParameters()
	if params != nil {
		res["QueryParameters"] = EntityToJson_convertEntityToJson(params, nil)
	} else {
		res["QueryParameters"] = nil
	}
	return res
}

func JsonExtensions_tryGetConflict(metadata ObjectNode) bool {
	v, ok := metadata[Constants_Documents_Metadata_CONFLICT]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}
