package ravendb

import (
	"encoding/json"
	"fmt"
	"io"
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

func isDelimToken(tok json.Token, delim string) bool {
	delimTok, ok := tok.(json.Delim)
	return ok && delimTok.String() == delim
}

func (o *StreamOperation) setResult(response *StreamResultResponse) (*YieldStreamResults, error) {
	if response == nil {
		return nil, NewIllegalStateException("The index does not exists, failed to stream results")
	}
	dec := json.NewDecoder(response.Stream)
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	// we expect start of array token
	if !isDelimToken(tok, "[") {
		return nil, NewIllegalStateException("Expected start object ', got %T %s", tok, tok)
	}

	if o._isQueryStream {
		o._statistics, err = handleStreamQueryStats(dec)
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
		return nil, NewIllegalStateException("Expected start object ', got %T %s", tok, tok)
	}

	return NewYieldStreamResults(response, dec), nil
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

func handleStreamQueryStats(dec *json.Decoder) (*StreamQueryStatistics, error) {
	var stats StreamQueryStatistics
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
			stats.IndexTimestamp, err = NetISO8601Utils_parse(s)
		}
	}
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

type YieldStreamResults struct {
	response *StreamResultResponse
	dec      *json.Decoder
	err      error
}

func NewYieldStreamResults(response *StreamResultResponse, dec *json.Decoder) *YieldStreamResults {
	return &YieldStreamResults{
		response: response,
		dec:      dec,
	}
}

// decodes next javascript object from stream
// returns io.EOF when reaching end of stream. Other errors indicate a parsing error
func (r *YieldStreamResults) Next() (ObjectNode, error) {
	if r.err != nil {
		return nil, r.err
	}
	// More() returns false if there is an error or ']' token
	if r.dec.More() {
		var v ObjectNode
		r.err = r.dec.Decode(&v)
		if r.err != nil {
			return nil, r.err
		}
		return v, nil
	}
	// expect end of Results array
	r.eatArrayEnd()
	if r.err != nil {
		return nil, r.err
	}

	// should now return nil, io.EOF to indicate end of stream
	_, r.err = r.dec.Token()
	return nil, r.err
}

func (r *YieldStreamResults) eatArrayEnd() {
	if r.err != nil {
		return
	}
	var tok json.Token
	tok, r.err = r.dec.Token()
	if r.err == nil && !isDelimToken(tok, "]") {
		r.err = fmt.Errorf("Expected ']' token. Got token of type %T, value: '%s'", tok, tok)
	}
}

func (r *YieldStreamResults) Close() {
	// a bit of a hack
	if rc, ok := r.response.Stream.(io.ReadCloser); ok {
		rc.Close()
	}
}
