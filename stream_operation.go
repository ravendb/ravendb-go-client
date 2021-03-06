package ravendb

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// StreamOperation represents a streaming operation
type StreamOperation struct {
	session       *InMemoryDocumentSessionOperations
	statistics    *StreamQueryStatistics
	isQueryStream bool
}

// NewStreamOperation returns new StreamOperation
func NewStreamOperation(session *InMemoryDocumentSessionOperations, statistics *StreamQueryStatistics) *StreamOperation {
	return &StreamOperation{
		session:    session,
		statistics: statistics,
	}
}

func (o *StreamOperation) createRequestForIndexQuery(query *IndexQuery) (*QueryStreamCommand, error) {
	o.isQueryStream = true

	if query.waitForNonStaleResults {
		return nil, newUnsupportedOperationError("Since stream() does not wait for indexing (by design), streaming query with setWaitForNonStaleResults is not supported")
	}

	if err := o.session.incrementRequestCount(); err != nil {
		return nil, err
	}

	return NewQueryStreamCommand(o.session.Conventions, query), nil
}

func (o *StreamOperation) createRequest(startsWith string, matches string, start int, pageSize int, exclude string, startAfter string) *StreamCommand {
	uri := "streams/docs?"

	if startsWith != "" {
		uri += "startsWith=" + urlUtilsEscapeDataString(startsWith) + "&"
	}

	if matches != "" {
		uri += "matches=" + urlUtilsEscapeDataString(matches) + "&"
	}

	if exclude != "" {
		uri += "exclude=" + urlUtilsEscapeDataString(exclude) + "&"
	}

	if startAfter != "" {
		uri += "startAfter=" + urlUtilsEscapeDataString(startAfter) + "&"
	}

	if start != 0 {
		uri += "start=" + strconv.Itoa(start) + "&"
	}

	// Note: using 0 as default value instead of MaxInt
	if pageSize != 0 {
		uri += "pageSize=" + strconv.Itoa(pageSize) + "&"
	}

	uri = strings.TrimSuffix(uri, "&")
	return NewStreamCommand(uri)
}

func isDelimToken(tok json.Token, delim string) bool {
	delimTok, ok := tok.(json.Delim)
	return ok && delimTok.String() == delim
}

/* The response looks like:
{
  "Results": [
    {
       "foo": bar,
    }
  ]
}
*/
func (o *StreamOperation) setResult(response *StreamResultResponse) (*yieldStreamResults, error) {
	if response == nil {
		return nil, newIllegalStateError("The index does not exists, failed to stream results")
	}
	dec := json.NewDecoder(response.Stream)
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	// we expect start of json object
	if !isDelimToken(tok, "{") {
		return nil, newIllegalStateError("Expected start object '{', got %T %s", tok, tok)
	}

	if o.isQueryStream {
		if o.statistics == nil {
			o.statistics = &StreamQueryStatistics{}
		}
		err = handleStreamQueryStats(dec, o.statistics)
		if err != nil {
			return nil, err
		}
	}

	// expecting object with a single field "Results" that is array of values
	tok, err = getTokenAfterObjectKey(dec, "Results")
	if err != nil {
		return nil, err
	}
	if !isDelimToken(tok, "[") {
		return nil, newIllegalStateError("Expected start array '[', got %T %s", tok, tok)
	}

	return newYieldStreamResults(response, dec), nil
}

func getNextDelimToken(dec *json.Decoder, delimStr string) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); ok || delim.String() == delimStr {
		return nil
	}
	return fmt.Errorf("Expected delim token '%s', got %T %s", delimStr, tok, tok)
}

func getNextStringToken(dec *json.Decoder) (string, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", err
	}
	if s, ok := tok.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("Expected string token, got %T %s", tok, tok)
}

func getTokenAfterObjectKey(dec *json.Decoder, name string) (json.Token, error) {
	s, err := getNextStringToken(dec)
	if err == nil {
		if s != name {
			return nil, fmt.Errorf("Expected string token named '%s', got '%s'", name, s)
		}
	}
	return dec.Token()
}

func getNextObjectStringValue(dec *json.Decoder, name string) (string, error) {
	tok, err := getTokenAfterObjectKey(dec, name)
	if err != nil {
		return "", err
	}
	s, ok := tok.(string)
	if !ok {
		return "", fmt.Errorf("Expected string token, got %T %s", tok, tok)
	}
	return s, nil
}

func getNextObjectBoolValue(dec *json.Decoder, name string) (bool, error) {
	tok, err := getTokenAfterObjectKey(dec, name)
	if err != nil {
		return false, err
	}
	v, ok := tok.(bool)
	if !ok {
		return false, fmt.Errorf("Expected bool token, got %T %s", tok, tok)
	}
	return v, nil
}

func getNextObjectInt64Value(dec *json.Decoder, name string) (int64, error) {
	tok, err := getTokenAfterObjectKey(dec, name)
	if err != nil {
		return 0, err
	}
	if v, ok := tok.(float64); ok {
		return int64(v), nil
	}
	if v, ok := tok.(json.Number); ok {
		return v.Int64()
	}
	return 0, fmt.Errorf("Expected number token, got %T %s", tok, tok)
}

func handleStreamQueryStats(dec *json.Decoder, stats *StreamQueryStatistics) error {
	var err error
	var n int64
	stats.ResultEtag, err = getNextObjectInt64Value(dec, "ResultEtag")
	if err == nil {
		stats.IsStale, err = getNextObjectBoolValue(dec, "IsStale")
	}
	if err == nil {
		stats.IndexName, err = getNextObjectStringValue(dec, "IndexName")
	}
	if err == nil {
		n, err = getNextObjectInt64Value(dec, "TotalResults")
		stats.TotalResults = int(n)
	}
	if err == nil {
		var s string
		s, err = getNextObjectStringValue(dec, "IndexTimestamp")
		if err == nil {
			stats.IndexTimestamp, err = ParseTime(s)
		}
	}
	return err
}

type yieldStreamResults struct {
	response *StreamResultResponse
	dec      *json.Decoder
	err      error
}

func newYieldStreamResults(response *StreamResultResponse, dec *json.Decoder) *yieldStreamResults {
	return &yieldStreamResults{
		response: response,
		dec:      dec,
	}
}

// next decodes next value from stream
// returns io.EOF when reaching end of stream. Other errors indicate a parsing error
func (r *yieldStreamResults) next(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	// More() returns false if there is an error or ']' token
	if r.dec.More() {
		r.err = r.dec.Decode(&v)
		if r.err != nil {
			return r.err
		}
		return nil
	}

	// expect end of Results array
	r.err = getNextDelimToken(r.dec, "]")
	if r.err != nil {
		return r.err
	}

	// expect end of top-level json object
	r.err = getNextDelimToken(r.dec, "}")
	if r.err != nil {
		return r.err
	}

	// should now return nil, io.EOF to indicate end of stream
	_, r.err = r.dec.Token()
	return r.err
}

// nextJSONObject decodes next javascript object from stream
// returns io.EOF when reaching end of stream. Other errors indicate a parsing error
func (r *yieldStreamResults) nextJSONObject() (map[string]interface{}, error) {
	var v map[string]interface{}
	err := r.next(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *yieldStreamResults) close() error {
	// a bit of a hack
	if rc, ok := r.response.Stream.(io.ReadCloser); ok {
		return rc.Close()
	}
	return nil
}
