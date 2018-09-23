package ravendb

import (
	"math"
	"strconv"
	"strings"
)

type StreamOperation struct {
	_session       *InMemoryDocumentSessionOperations
	_statistics    *StreamQueryStatistics
	_isQueryStream bool
}

func NewStreamOperation(session *InMemoryDocumentSessionOperations, statistics *StreamQueryStatistics) *StreamOperation {
	return &StreamOperation{
		_session:    session,
		_statistics: statistics,
	}
}

func (o *StreamOperation) createRequest(query *IndexQuery) *QueryStreamCommand {
	o._isQueryStream = true

	if query.waitForNonStaleResults {
		//throw new UnsupportedOperationException("Since stream() does not wait for indexing (by design), streaming query with setWaitForNonStaleResults is not supported");
		panic("Since stream() does not wait for indexing (by design), streaming query with setWaitForNonStaleResults is not supported")
	}

	o._session.IncrementRequestCount()

	return NewQueryStreamCommand(o._session.Conventions, query)
}

func (o *StreamOperation) createRequest2(startsWith string, matches string, start int, pageSize int, exclude string, startAfter string) *StreamCommand {
	uri := "streams/docs?"

	if startsWith != "" {
		uri += "startsWith=" + UrlUtils_escapeDataString(startsWith) + "&"
	}

	if matches != "" {
		uri += "matches=" + UrlUtils_escapeDataString(matches) + "&"
	}

	if exclude != "" {
		uri += "exclude=" + UrlUtils_escapeDataString(exclude) + "&"
	}

	if startAfter != "" {
		uri += "startAfter=" + UrlUtils_escapeDataString(startAfter) + "&"
	}

	if start != 0 {
		uri += "start=" + strconv.Itoa(start) + "&"
	}

	// TODO: verify callers use this number. Maybe can use 0 instead?
	if pageSize != math.MaxInt32 {
		uri += "pageSize=" + strconv.Itoa(pageSize) + "&"
	}

	uri = strings.TrimSuffix(uri, "&")
	return NewStreamCommand(uri)
}

/*
   public CloseableIterator<ObjectNode> setResult(StreamResultResponse response)  {
       if (response == null) {
           throw new IllegalStateException("The index does not exists, failed to stream results");
       }

       try {
           JsonParser parser = JsonExtensions.getDefaultMapper().getFactory().createParser(response.getStream());

           if (parser.nextToken() != JsonToken.START_OBJECT) {
               throw new IllegalStateException("Expected start object");
           }

           if (_isQueryStream) {
               handleStreamQueryStats(parser, _statistics);
           }

           if (!"Results".equals(parser.nextFieldName())) {
               throw new IllegalStateException("Expected Results field");
           }

           if (parser.nextToken() != JsonToken.START_ARRAY) {
               throw new IllegalStateException("Expected results array start");
           }

           return new YieldStreamResults(response, parser);
       } catch (IOException e) {
           throw new RuntimeException("Unable to stream result: " + e.getMessage(), e);
       }
   }

   private static void handleStreamQueryStats(JsonParser parser, StreamQueryStatistics streamQueryStatistics) throws IOException {
       if (!"ResultEtag".equals(parser.nextFieldName())) {
           throw new IllegalStateException("Expected ResultETag field");
       }

       long resultEtag = parser.nextLongValue(0);

       if (!"IsStale".equals(parser.nextFieldName())) {
           throw new IllegalStateException("Expected IsStale field");
       }

       boolean isStale = parser.nextBooleanValue();

       if (!"IndexName".equals(parser.nextFieldName())) {
           throw new IllegalStateException("Expected IndexName field");
       }

       String indexName = parser.nextTextValue();

       if (!"TotalResults".equals(parser.nextFieldName())) {
           throw new IllegalStateException("Expected TotalResults field");
       }

       int totalResults = (int) parser.nextLongValue(0);

       if (!"IndexTimestamp".equals(parser.nextFieldName())) {
           throw new IllegalStateException("Expected IndexTimestamp field");
       }

       String indexTimestamp = parser.nextTextValue();

       if (streamQueryStatistics == null) {
           return;
       }

       streamQueryStatistics.setIndexName(indexName);
       streamQueryStatistics.setStale(isStale);
       streamQueryStatistics.setTotalResults(totalResults);
       streamQueryStatistics.setResultEtag(resultEtag);
       streamQueryStatistics.setIndexTimestamp(NetISO8601Utils.parse(indexTimestamp));
   }

   private class YieldStreamResults implements CloseableIterator<ObjectNode> {

       private StreamResultResponse response;
       private JsonParser parser;

       public YieldStreamResults(StreamResultResponse response, JsonParser parser) {
           this.response = response;
           this.parser = parser;
       }

       @Override
       public ObjectNode next() {
           try {
               ObjectNode node = JsonExtensions.getDefaultMapper().readTree(parser);
               return node;
           } catch (IOException e) {
               throw new IllegalStateException("Unable to read stream result: " + e.getMessage(), e);
           }
       }

       @Override
       public boolean hasNext() {
           try {
               JsonToken jsonToken = parser.nextToken();
               if (jsonToken == JsonToken.END_ARRAY) {

                   if (parser.nextToken() != JsonToken.END_OBJECT) {
                       throw new IllegalStateException("Expected '}' after results array");
                   }

                   return false;
               }

               return true;
           } catch (IOException e) {
               throw new IllegalStateException("Unable to read stream result: " + e.getMessage(), e);
           }
       }

       @Override
       public void close() {
           try {
               response.getResponse().close();
           } catch (IOException e) {
               throw new RuntimeException("Unable to close stream response");
           }
       }


   }
*/
