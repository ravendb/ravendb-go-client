package ravendb

import (
	"fmt"
	"reflect"
)

type QueryOperation struct {
	_session                 *InMemoryDocumentSessionOperations
	_indexName               string
	_indexQuery              *IndexQuery
	_metadataOnly            bool
	_indexEntriesOnly        bool
	_currentQueryResults     *QueryResult
	_fieldsToFetch           *FieldsToFetchToken
	_sp                      *Stopwatch
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

func (o *QueryOperation) CreateRequest() *QueryCommand {
	o._session.IncrementRequestCount()

	//o.logQuery();

	return NewQueryCommand(o._session.GetConventions(), o._indexQuery, o._metadataOnly, o._indexEntriesOnly)
}

func (o *QueryOperation) getCurrentQueryResults() *QueryResult {
	return o._currentQueryResults
}

func (o *QueryOperation) setResult(queryResult *QueryResult) {
	o.ensureIsAcceptableAndSaveResult(queryResult)
}

func (o *QueryOperation) assertPageSizeSet() {
	if !o._session.GetConventions().IsThrowIfQueryPageSizeIsNotSet() {
		return
	}

	if o._indexQuery.pageSize > 0 {
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

	if !o._indexQuery.waitForNonStaleResults {
		var res *NilCleanCloseable
		return res
	}

	return o._session.GetDocumentStore().DisableAggressiveCachingWithDatabase(o._session.GetDatabaseName())
}

func (o *QueryOperation) completeNew(results interface{}) error {
	queryResult := o._currentQueryResults.createSnapshot()

	if !o._disableEntitiesTracking {
		o._session.RegisterIncludes(queryResult.getIncludes())
	}
	rt := reflect.TypeOf(results)

	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("results should be a pointer to a slice of pointers to struct, is %T. rt: %s", results, rt)
	}
	rv := reflect.ValueOf(results)
	sliceV := rv.Elem()

	// slice element should be a pointer to a struct
	sliceElemPtrType := sliceV.Type().Elem()
	if sliceElemPtrType.Kind() != reflect.Ptr {
		return fmt.Errorf("results should be a pointer to a slice of pointers to struct, is %T. sliceElemPtrType: %s", results, sliceElemPtrType)
	}

	sliceElemType := sliceElemPtrType.Elem()
	if sliceElemType.Kind() != reflect.Struct {
		return fmt.Errorf("results should be a pointer to a slice of pointers to struct, is %T. sliceElemType: %s", results, sliceElemType)
	}
	// if this is a pointer to nil slice, create a new slice
	// otherwise we use the slice that was provided by the caller
	if sliceV.IsNil() {
		sliceV.Set(reflect.MakeSlice(sliceV.Type(), 0, 0))
	}

	sliceV2 := sliceV

	clazz := sliceElemPtrType
	for _, document := range queryResult.Results {
		metadataI, ok := document[Constants_Documents_Metadata_KEY]
		panicIf(!ok, "missing metadata")
		metadata := metadataI.(ObjectNode)
		id, _ := JsonGetAsText(metadata, Constants_Documents_Metadata_ID)

		el, err := QueryOperation_deserialize(clazz, id, document, metadata, o._fieldsToFetch, o._disableEntitiesTracking, o._session)
		if err != nil {
			return NewRuntimeException("Unable to read json: %s", err)
		}
		v2 := reflect.ValueOf(el)
		sliceV2 = reflect.Append(sliceV2, v2)
	}

	if !o._disableEntitiesTracking {
		o._session.RegisterMissingIncludes(queryResult.Results, queryResult.getIncludes(), queryResult.getIncludedPaths())
	}
	if sliceV2 != sliceV {
		sliceV.Set(sliceV2)
	}
	return nil
}

func (o *QueryOperation) completeOld(clazz reflect.Type) ([]interface{}, error) {
	queryResult := o._currentQueryResults.createSnapshot()

	if !o._disableEntitiesTracking {
		o._session.RegisterIncludes(queryResult.getIncludes())
	}

	var list []interface{}
	{
		results := queryResult.getResults()
		for _, document := range results {
			metadataI, ok := document[Constants_Documents_Metadata_KEY]
			panicIf(!ok, "missing metadata")
			metadata := metadataI.(ObjectNode)
			id, _ := JsonGetAsText(metadata, Constants_Documents_Metadata_ID)
			el, err := QueryOperation_deserialize(clazz, id, document, metadata, o._fieldsToFetch, o._disableEntitiesTracking, o._session)
			if err != nil {
				return nil, NewRuntimeException("Unable to read json: %s", err)
			}
			list = append(list, el)
		}
	}

	if !o._disableEntitiesTracking {
		o._session.RegisterMissingIncludes(queryResult.getResults(), queryResult.getIncludes(), queryResult.getIncludedPaths())
	}

	return list, nil
}

func jsonIsValueNode(v interface{}) bool {
	switch v.(type) {
	case string, float64, bool:
		return true
	case []interface{}, ObjectNode:
		return false
	}
	panicIf(true, "unhandled type %T", v)
	return false
}

func QueryOperation_deserialize(clazz reflect.Type, id string, document ObjectNode, metadata ObjectNode, fieldsToFetch *FieldsToFetchToken, disableEntitiesTracking bool, session *InMemoryDocumentSessionOperations) (interface{}, error) {
	_, ok := jsonGetAsBool(metadata, "@projection")
	if !ok {
		return session.TrackEntityOld(clazz, id, document, metadata, disableEntitiesTracking)
	}
	if fieldsToFetch != nil && len(fieldsToFetch.projections) == 1 {
		// we only select a single field
		isString := clazz.Kind() == reflect.String
		if isString || ClassUtils_isPrimitiveOrWrapper(clazz) || typeIsEnum(clazz) {
			projectField := fieldsToFetch.projections[0]
			jsonNode, ok := document[projectField]
			if ok && jsonIsValueNode(jsonNode) {
				res, err := session.GetConventions().DeserializeEntityFromJson(clazz, jsonNode)
				if err != nil {
					return nil, err
				}
				if res != nil {
					return res, nil
				}
				return Defaults_defaultValue(clazz), nil
			}
		}

		inner, ok := document[fieldsToFetch.projections[0]]
		if !ok {
			return Defaults_defaultValue(clazz), nil
		}

		if fieldsToFetch.fieldsToFetch != nil && fieldsToFetch.fieldsToFetch[0] == fieldsToFetch.projections[0] {
			doc, ok := inner.(ObjectNode)
			if ok {
				//extraction from original type
				document = doc
			}
		}
	}

	result, err := treeToValue(clazz, document)
	if err != nil {
		return nil, err
	}

	if StringUtils_isNotEmpty(id) {
		// we need to make an additional check, since it is possible that a value was explicitly stated
		// for the identity property, in which case we don't want to override it.

		identityProperty := session.GetConventions().GetIdentityProperty(clazz)
		if identityProperty != "" {
			if _, ok := document[identityProperty]; !ok {
				session.GetGenerateEntityIdOnTheClient().trySetIdentity(result, id)
			}
		}
	}

	return result, nil
}

func (o *QueryOperation) isDisableEntitiesTracking() bool {
	return o._disableEntitiesTracking
}

func (o *QueryOperation) setDisableEntitiesTracking(disableEntitiesTracking bool) {
	o._disableEntitiesTracking = disableEntitiesTracking
}

func (o *QueryOperation) ensureIsAcceptableAndSaveResult(result *QueryResult) error {
	if result == nil {
		return NewIndexDoesNotExistException("Could not find index " + o._indexName)
	}

	err := QueryOperation_ensureIsAcceptable(result, o._indexQuery.waitForNonStaleResults, o._sp, o._session)
	if err != nil {
		return err
	}
	o._currentQueryResults = result

	// TODO: port me when we have logger
	/*
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
	*/
	return nil
}

func QueryOperation_ensureIsAcceptable(result *QueryResult, waitForNonStaleResults bool, duration *Stopwatch, session *InMemoryDocumentSessionOperations) error {
	if waitForNonStaleResults && result.isStale() {
		duration.stop()
		msg := "Waited for " + duration.String() + " for the query to return non stale result."
		return NewTimeoutException(msg)
	}
	return nil
}

func (o *QueryOperation) getIndexQuery() *IndexQuery {
	return o._indexQuery
}
