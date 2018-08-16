package ravendb

type QueryResult struct {
	GenericQueryResult
}

func (r *QueryResult) createSnapshot() *QueryResult {
	queryResult := *r

	/* TBD 4.1
	Map<String, Map<String, List<String>>> highlightings = getHighlightings();

	if (highlightings != null) {
		Map<String, Map<String, List<String>>> newHighlights = new HashMap<>();
		for (Map.Entry<String, Map<String, List<String>>> hightlightEntry : getHighlightings().entrySet()) {
			newHighlights.put(hightlightEntry.getKey(), new HashMap<>(hightlightEntry.getValue()));
		}
		queryResult.setHighlightings(highlightings);
	}*/

	queryResult.ScoreExplanations = dupMapStringString(r.ScoreExplanations)
	queryResult.TimingsInMs = dupMapStringFloat64(r.TimingsInMs)
	return &queryResult
}
