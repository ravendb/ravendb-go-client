package ravendb

import (
	"fmt"
	"io"
	"reflect"
)

// QueryOperation describes query operation
type QueryOperation struct {
	session                 *InMemoryDocumentSessionOperations
	indexName               string
	indexQuery              *IndexQuery
	metadataOnly            bool
	indexEntriesOnly        bool
	currentQueryResults     *QueryResult
	fieldsToFetch           *fieldsToFetchToken
	sp                      *stopWatch
	disableEntitiesTracking bool

	// static  Log logger = LogFactory.getLog(QueryOperation.class);
}

// NewQueryOperation returns new QueryOperation
func NewQueryOperation(session *InMemoryDocumentSessionOperations, indexName string, indexQuery *IndexQuery, fieldsToFetch *fieldsToFetchToken, disableEntitiesTracking bool, metadataOnly bool, indexEntriesOnly bool) (*QueryOperation, error) {
	res := &QueryOperation{
		session:                 session,
		indexName:               indexName,
		indexQuery:              indexQuery,
		fieldsToFetch:           fieldsToFetch,
		disableEntitiesTracking: disableEntitiesTracking,
		metadataOnly:            metadataOnly,
		indexEntriesOnly:        indexEntriesOnly,
	}
	if err := res.assertPageSizeSet(); err != nil {
		return nil, err
	}
	return res, nil
}

// CreateRequest creates a request
func (o *QueryOperation) CreateRequest() (*QueryCommand, error) {
	if err := o.session.incrementRequestCount(); err != nil {
		return nil, err
	}

	//o.logQuery();

	return NewQueryCommand(o.session.GetConventions(), o.indexQuery, o.metadataOnly, o.indexEntriesOnly)
}

func (o *QueryOperation) setResult(queryResult *QueryResult) {
	o.ensureIsAcceptableAndSaveResult(queryResult)
}

func (o *QueryOperation) assertPageSizeSet() error {
	if !o.session.GetConventions().IsThrowIfQueryPageSizeIsNotSet() {
		return nil
	}

	if o.indexQuery.pageSize > 0 {
		return nil
	}

	return newIllegalStateError("Attempt to query without explicitly specifying a page size. " +
		"You can use .take() methods to set maximum number of results. By default the page //size is set to Integer.MAX_VALUE and can cause severe performance degradation.")
}

func (o *QueryOperation) startTiming() {
	o.sp = newStopWatchStarted()
}

func (o *QueryOperation) logQuery() {
	/*
		if (logger.isInfoEnabled()) {
			logger.info("Executing query " + _indexQuery.getQuery() + " on index " + _indexName + " in " + _session.storeIdentifier());
		}
	*/
}

func (o *QueryOperation) enterQueryContext() io.Closer {
	o.startTiming()

	if !o.indexQuery.waitForNonStaleResults {
		var res *nilCloser
		return res
	}

	return o.session.GetDocumentStore().DisableAggressiveCachingWithDatabase(o.session.DatabaseName)
}

// results is *[]<type> and we'll create the slice and fill it with values
// of <type> and do the equivalent of: *results = our_slice
func (o *QueryOperation) complete(results interface{}) error {
	queryResult := o.currentQueryResults.createSnapshot()

	if !o.disableEntitiesTracking {
		o.session.RegisterIncludes(queryResult.Includes)
	}
	rt := reflect.TypeOf(results)

	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("results should *[]<type>, is %T. rt: %s", results, rt)
	}
	rv := reflect.ValueOf(results)
	sliceV := rv.Elem()
	// if this is a pointer to nil slice, create a new slice
	// otherwise we use the slice that was provided by the caller
	// TODO: should this always be a new slice? (in which case we should error
	// if provided non-nil slice, since that implies user error
	// r at least we should reset the slice to empty. Appending to existing
	// slice might be confusing/unexpected to callers
	if sliceV.IsNil() {
		sliceV.Set(reflect.MakeSlice(sliceV.Type(), 0, 0))
	}

	sliceV2 := sliceV

	clazz := sliceV.Type().Elem()
	for _, document := range queryResult.Results {
		metadataI, ok := document[MetadataKey]
		panicIf(!ok, "missing metadata")
		metadata := metadataI.(map[string]interface{})
		id, _ := jsonGetAsText(metadata, MetadataID)
		result := reflect.New(clazz) // this is a pointer to desired value
		err := queryOperationDeserialize(result.Interface(), id, document, metadata, o.fieldsToFetch, o.disableEntitiesTracking, o.session)
		if err != nil {
			return newRuntimeError("Unable to read json: %s", err)
		}
		// de-reference pointer value
		v2 := result.Elem()
		sliceV2 = reflect.Append(sliceV2, v2)
	}

	if !o.disableEntitiesTracking {
		o.session.RegisterMissingIncludes(queryResult.Results, queryResult.Includes, queryResult.IncludedPaths)
	}
	if sliceV2 != sliceV {
		sliceV.Set(sliceV2)
	}
	return nil
}

func jsonIsValueNode(v interface{}) bool {
	switch v.(type) {
	case string, float64, bool:
		return true
	case []interface{}, map[string]interface{}:
		return false
	}
	panicIf(true, "unhandled type %T", v)
	return false
}

// result is pointer to value that will be set with value decoded from JSON
func queryOperationDeserialize(result interface{}, id string, document map[string]interface{}, metadata map[string]interface{}, fieldsToFetch *fieldsToFetchToken, disableEntitiesTracking bool, session *InMemoryDocumentSessionOperations) error {
	_, ok := jsonGetAsBool(metadata, MetadataProjection)
	if !ok {
		return session.TrackEntity(result, id, document, metadata, disableEntitiesTracking)
	}
	tp := reflect.TypeOf(result)
	panicIf(tp.Kind() != reflect.Ptr, "result should be a *<type>, is %T", result)
	clazz := tp.Elem()
	if fieldsToFetch != nil && len(fieldsToFetch.projections) == 1 {
		// we only select a single field
		isString := clazz.Kind() == reflect.String
		if isString || isPrimitiveOrWrapper(clazz) || typeIsEnum(clazz) {
			projectField := fieldsToFetch.projections[0]
			jsonNode, ok := document[projectField]
			if ok && jsonIsValueNode(jsonNode) {
				res, err := treeToValue(clazz, jsonNode)
				if err != nil {
					return err
				}
				if res != nil {
					setInterfaceToValue(result, res)
					return nil
				}
				return nil
			}
		}

		inner, ok := document[fieldsToFetch.projections[0]]
		if !ok {
			return nil
		}

		if fieldsToFetch.fieldsToFetch != nil && fieldsToFetch.fieldsToFetch[0] == fieldsToFetch.projections[0] {
			doc, ok := inner.(map[string]interface{})
			if ok {
				// extraction from original type
				document = doc
			}
		}
	}

	res, err := treeToValue(clazz, document)
	if err != nil {
		return err
	}

	if stringIsNotEmpty(id) {
		// we need to make an additional check, since it is possible that a value was explicitly stated
		// for the identity property, in which case we don't want to override it.

		identityProperty := session.GetConventions().GetIdentityProperty(clazz)
		if identityProperty != "" {
			if _, ok := document[identityProperty]; !ok {
				session.GetGenerateEntityIDOnTheClient().trySetIdentity(res, id)
			}
		}
	}

	setInterfaceToValue(result, res)
	return nil
}

func (o *QueryOperation) ensureIsAcceptableAndSaveResult(result *QueryResult) error {
	if result == nil {
		return newIndexDoesNotExistError("Could not find index " + o.indexName)
	}

	err := queryOperationEnsureIsAcceptable(result, o.indexQuery.waitForNonStaleResults, o.sp, o.session)
	if err != nil {
		return err
	}
	o.currentQueryResults = result

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

func queryOperationEnsureIsAcceptable(result *QueryResult, waitForNonStaleResults bool, duration *stopWatch, session *InMemoryDocumentSessionOperations) error {
	if waitForNonStaleResults && result.IsStale {
		duration.stop()
		msg := "Waited for " + duration.String() + " for the query to return non stale result."
		return NewTimeoutError(msg)
	}
	return nil
}
