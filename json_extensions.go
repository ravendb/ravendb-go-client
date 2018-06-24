package ravendb

// TODO: change the name to better reflect what it does
func JsonExtensions_writeIndexQuery(conventions *DocumentConventions, query *IndexQuery) map[string]interface{} {
	res := map[string]interface{}{}
	res["Query"] = query.getQuery()
	if query.isPageSizeSet() && query.getPageSize() >= 0 {
		res["PageSize"] = query.getPageSize()
	}

	if query.isWaitForNonStaleResults() {
		res["WaitForNonStaleResults"] = query.isWaitForNonStaleResults()
	}

	if query.getStart() > 0 {
		res["Start"] = query.getStart()
	}

	if query.getWaitForNonStaleResultsTimeout() != 0 {
		s := TimeUtils_durationToTimeSpan(query.getWaitForNonStaleResultsTimeout())
		res["WaitForNonStaleResultsTimeout"] = s
	}

	if query.isDisableCaching() {
		res["DisableCaching"] = query.isDisableCaching()
	}

	if query.isSkipDuplicateChecking() {
		res["SkipDuplicateChecking"] = query.isSkipDuplicateChecking()
	}
	params := query.getQueryParameters()
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
