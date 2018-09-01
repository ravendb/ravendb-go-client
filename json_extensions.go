package ravendb

// TODO: change the name to better reflect what it does
func JsonExtensions_writeIndexQuery(conventions *DocumentConventions, query *IndexQuery) map[string]interface{} {
	res := map[string]interface{}{}
	res["Query"] = query.query
	if query.pageSize > 0 {
		res["PageSize"] = query.pageSize
	}

	if query.waitForNonStaleResults {
		res["WaitForNonStaleResults"] = query.waitForNonStaleResults
	}

	if query.start > 0 {
		res["Start"] = query.start
	}

	if query.waitForNonStaleResultsTimeout != 0 {
		s := TimeUtils_durationToTimeSpan(query.waitForNonStaleResultsTimeout)
		res["WaitForNonStaleResultsTimeout"] = s
	}

	if query.disableCaching {
		res["DisableCaching"] = query.disableCaching
	}

	if query.skipDuplicateChecking {
		res["SkipDuplicateChecking"] = query.skipDuplicateChecking
	}
	params := query.queryParameters
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
