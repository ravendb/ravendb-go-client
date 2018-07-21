package ravendb

// TODO: implement me
type QueryOperation struct {
	_session                 *InMemoryDocumentSessionOperations
	_indexName               string
	_indexQuery              *IndexQuery
	_metadataOnly            bool
	_indexEntriesOnly        bool
	_currentQueryResults     *QueryResult
	_fieldsToFetch           *FieldsToFetchToken
	_sp                      Stopwatch
	_disableEntitiesTracking bool

	// static  Log logger = LogFactory.getLog(QueryOperation.class);
}

func NewQueryOperation(session *InMemoryDocumentSessionOperations, indexName string, indexQuery *IndexQuery, fieldsToFetch *FieldsToFetchToken, disableEntitiesTracking bool, metadataOnly bool, indexEntriesOnly bool) *QueryOperation {
	res := &QueryOperation{
		_session:                 session,
		_indexName:               indexName,
		_indexQuery:              indexQuery,
		_fieldsToFetch:           fieldsToFetch,
		_disableEntitiesTracking: disableEntitiesTracking,
		_metadataOnly:            metadataOnly,
		_indexEntriesOnly:        indexEntriesOnly,
	}
	//res.assertPageSizeSet()
	return res
}

func (o *QueryOperation) createRequest() *QueryCommand {
	o._session.incrementRequestCount()

	//o.logQuery();

	return NewQueryCommand(o._session.getConventions(), o._indexQuery, o._metadataOnly, o._indexEntriesOnly)
}

func (o *QueryOperation) getCurrentQueryResults() *QueryResult {
	return o._currentQueryResults
}

func (o *QueryOperation) setResult(queryResult *QueryResult) {
	//o.ensureIsAcceptableAndSaveResult(queryResult)
}

func (o *QueryOperation) assertPageSizeSet() {
	if !o._session.getConventions().isThrowIfQueryPageSizeIsNotSet() {
		return
	}

	if o._indexQuery.isPageSizeSet() {
		return
	}

	//throw new IllegalStateException("Attempt to query without explicitly specifying a page size. " +
	//		"You can use .take() methods to set maximum number of results. By default the page //size is set to Integer.MAX_VALUE and can cause severe performance degradation.");
	panicIf(true, "Attempt to query without explicitly specifying a page size. "+
		"You can use .take() methods to set maximum number of results. By default the page size is set to Integer.MAX_VALUE and can cause severe performance degradation.")
}

func (o *QueryOperation) startTiming() {
	o._sp = Stopwatch_createStarted()
}

func (o *QueryOperation) logQuery() {
	/*
		if (logger.isInfoEnabled()) {
			logger.info("Executing query " + _indexQuery.getQuery() + " on index " + _indexName + " in " + _session.storeIdentifier());
		}
	*/
}

func (o *QueryOperation) enterQueryContext() CleanCloseable {
	o.startTiming()

	if !o._indexQuery.isWaitForNonStaleResults() {
		return nil
	}

	return o._session.getDocumentStore().disableAggressiveCachingWithDatabase(o._session.getDatabaseName())
}

/*
 class QueryOperation {


     <T> List<T> complete(Class<T> clazz) {
        QueryResult queryResult = _currentQueryResults.createSnapshot();

        if (!_disableEntitiesTracking) {
            _session.registerIncludes(queryResult.getIncludes());
        }

        ArrayList<T> list = new ArrayList<>();

        try {
            for (JsonNode document : queryResult.getResults()) {
                ObjectNode metadata = (ObjectNode) document.get(Constants.Documents.Metadata.KEY);
                JsonNode idNode = metadata.get(Constants.Documents.Metadata.ID);

                string id = null;
                if (idNode != null && idNode.isTextual()) {
                    id = idNode.asText();
                }

                list.add(deserialize(clazz, id, (ObjectNode) document, metadata, _fieldsToFetch, _disableEntitiesTracking, _session));
            }
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Unable to read json: " + e.getMessage(), e);
        }

        if (!_disableEntitiesTracking) {
            _session.registerMissingIncludes(queryResult.getResults(), queryResult.getIncludes(), queryResult.getIncludedPaths());
        }

        return list;
    }

    @SuppressWarnings("unchecked")
     static <T> T deserialize(Class<T> clazz, string id, ObjectNode document, ObjectNode metadata, FieldsToFetchToken fieldsToFetch, bool disableEntitiesTracking, InMemoryDocumentSessionOperations session) throws JsonProcessingException {

        JsonNode projection = metadata.get("@projection");
        if (projection == null || !projection.asBoolean()) {
            return (T)session.trackEntity(clazz, id, document, metadata, disableEntitiesTracking);
        }

        if (fieldsToFetch != null && fieldsToFetch.projections != null && fieldsToFetch.projections.length == 1) { // we only select a single field
            if (string.class.equals(clazz) || ClassUtils.isPrimitiveOrWrapper(clazz) || clazz.isEnum()) {
                string projectField = fieldsToFetch.projections[0];
                JsonNode jsonNode = document.get(projectField);
                if (jsonNode != null && jsonNode instanceof ValueNode) {
                    return ObjectUtils.firstNonNull((T) session.getConventions().deserializeEntityFromJson(clazz, jsonNode), Defaults.defaultValue(clazz));
                }
            }

            JsonNode inner = document.get(fieldsToFetch.projections[0]);
            if (inner == null) {
                return Defaults.defaultValue(clazz);
            }

            if (fieldsToFetch.fieldsToFetch != null && fieldsToFetch.fieldsToFetch[0].equals(fieldsToFetch.projections[0])) {
                if (inner instanceof ObjectNode) { //extraction from original type
                    document = (ObjectNode) inner;
                }
            }
        }

        T result = session.getConventions().getEntityMapper().treeToValue(document, clazz);

        if (stringUtils.isNotEmpty(id)) {
            // we need to make an additional check, since it is possible that a value was explicitly stated
            // for the identity property, in which case we don't want to override it.
            Field identityProperty = session.getConventions().getIdentityProperty(clazz);
            if (identityProperty != null) {
                JsonNode value = document.get(identityProperty.getName());

                if (value == null) {
                    session.getGenerateEntityIdOnTheClient().trySetIdentity(result, id);
                }
            }
        }

        return result;
    }

     bool isDisableEntitiesTracking() {
        return _disableEntitiesTracking;
    }

      setDisableEntitiesTracking(bool disableEntitiesTracking) {
        this._disableEntitiesTracking = disableEntitiesTracking;
    }

      ensureIsAcceptableAndSaveResult(QueryResult result) {
        if (result == null) {
            throw new IndexDoesNotExistException("Could not find index " + _indexName);
        }

        ensureIsAcceptable(result, _indexQuery.isWaitForNonStaleResults(), _sp, _session);

        _currentQueryResults = result;

        if (logger.isInfoEnabled()) {
            string isStale = result.isStale() ? " stale " : " ";

            stringBuilder parameters = new stringBuilder();
            if (_indexQuery.getQueryParameters() != null && !_indexQuery.getQueryParameters().isEmpty()) {
                parameters.append("(parameters: ");

                bool first = true;

                for (Map.Entry<string, Object> parameter : _indexQuery.getQueryParameters().entrySet()) {
                    if (!first) {
                        parameters.append(", ");
                    }

                    parameters.append(parameter.getKey())
                            .append(" = ")
                            .append(parameter.getValue());

                    first = false;
                }

                parameters.append(") ");
            }

            logger.info("Query " + _indexQuery.getQuery() + " " + parameters.tostring() + "returned " + result.getResults().size() + isStale + "results (total index results: " + result.getTotalResults() + ")");
        }
    }

     static  ensureIsAcceptable(QueryResult result, bool waitForNonStaleResults, Stopwatch duration, InMemoryDocumentSessionOperations session) {
        if (waitForNonStaleResults && result.isStale()) {
            duration.stop();

            string msg = "Waited for " + duration.tostring() + " for the query to return non stale result.";
            throw new TimeoutException(msg);

        }
    }


     IndexQuery getIndexQuery() {
        return _indexQuery;
    }
}
*/
