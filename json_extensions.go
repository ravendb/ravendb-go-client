package ravendb

func jsonExtensionsWriteIndexQuery(conventions *DocumentConventions, query *IndexQuery) map[string]interface{} {
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
		s := durationToTimeSpan(query.waitForNonStaleResultsTimeout)
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
		res["QueryParameters"] = convertEntityToJSON(params, nil)
	} else {
		res["QueryParameters"] = nil
	}
	return res
}

func jsonExtensionsTryGetConflict(metadata ObjectNode) bool {
	v, ok := metadata[MetadataConflict]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}
