package ravendb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Note: com.fasterxml.jackson.databind.JsonNode represents decoded JSON value i.e. interface{}

// Note: Java's com.fasterxml.jackson.databind.node.JsonNodeType represents type of JSON value, which in Go
// is the same as value itself, and therefore is interface{}

// Note: Java's com.fasterxml.jackson.databind.node.ObjectNode is map[string]interface{}
// It represents parsed json document

// Note: Java's com.fasterxml.jackson.databind.TreeNode represents a decoded JSON value that combines
// value and type. In Go it's interface{}

// Note: Java's com.fasterxml.jackson.databind.node.ArrayNode  represents array of JSON objects
// It's []map[string]interface{} in Go

// we should use jsonMarshal instead of jsonMarshal so that it's easy
// to change json marshalling in all code base (e.g. to use a faster
// json library or ensure that values are marshaled correctly)
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}

func jsonGetAsTextPointer(doc map[string]interface{}, key string) *string {
	v, ok := doc[key]
	if !ok {
		return nil
	}
	// TODO: only allow *string ?
	if s, ok := v.(*string); ok {
		return s
	}
	s := v.(string)
	return &s
}

func jsonGetAsString(doc map[string]interface{}, key string) (string, bool) {
	return jsonGetAsText(doc, key)
}

func jsonGetAsText(doc map[string]interface{}, key string) (string, bool) {
	v, ok := doc[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func jsonGetAsInt(doc map[string]interface{}, key string) (int, bool) {
	v, ok := doc[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	if ok {
		return int(f), true
	}
	s, ok := v.(string)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

func jsonGetAsInt64(doc map[string]interface{}, key string) (int64, bool) {
	v, ok := doc[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	if ok {
		return int64(f), true
	}
	s, ok := v.(string)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func jsonGetAsBool(doc map[string]interface{}, key string) (bool, bool) {
	v, ok := doc[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	if ok {
		return b, true
	}
	s, ok := v.(string)
	if !ok {
		return false, false
	}
	if strings.EqualFold(s, "true") {
		return true, true
	}
	if strings.EqualFold(s, "false") {
		return false, true
	}
	return false, false
}

// converts a struct to JSON representations as map of string to value
// TODO: could be faster
func structToJSONMap(v interface{}) map[string]interface{} {
	d, err := jsonMarshal(v)
	must(err)
	var res map[string]interface{}
	err = jsonUnmarshal(d, &res)
	must(err)
	return res
}

// given a json in the form of map[string]interface{}, de-serialize it to a struct
// TODO: could be faster
func structFromJSONMap(js map[string]interface{}, v interface{}) error {
	d, err := jsonMarshal(js)
	if err != nil {
		return err
	}
	return jsonUnmarshal(d, v)
}

// matches a Java naming from EnityMapper
func valueToTree(v interface{}) map[string]interface{} {
	return structToJSONMap(v)
}

// TODO: remove
// copyJSONMap makes a deep copy of map[string]interface{}
// TODO: possibly not the fastest way to do it
/*
func copyJSONMap(v map[string]interface{}) map[string]interface{} {
	d, err := jsonMarshal(v)
	must(err)
	var res map[string]interface{}
	err = jsonUnmarshal(d, &res)
	must(err)
	return res
}
*/

// jsonDecodeFirst decode first JSON object from d
// This is like jsonUnmarshal() but allows for d
// to contain multiple JSON objects
// This is for compatibility with Java's ObjectMapper.readTree()
func jsonUnmarshalFirst(d []byte, v interface{}) error {
	r := bytes.NewReader(d)
	dec := json.NewDecoder(r)
	err := dec.Decode(v)
	if err != nil {
		dbg("jsonDecodeFirst: dec.Decode() of type %T failed with %s. JSON:\n%s\n\n", v, err, string(d))
	}
	return err
}

func isUnprintable(c byte) bool {
	if c < 32 {
		// 9 - tab, 10 - LF, 13 - CR
		if c == 9 || c == 10 || c == 13 {
			return false
		}
		return true
	}
	return c >= 127
}

func isBinaryData(d []byte) bool {
	for _, b := range d {
		if isUnprintable(b) {
			return true
		}
	}
	return false
}

func asHex(d []byte) ([]byte, bool) {
	if !isBinaryData(d) {
		return d, false
	}

	// convert unprintable characters to hex
	var res []byte
	for i, c := range d {
		if i > 2048 {
			break
		}
		if isUnprintable(c) {
			s := fmt.Sprintf("x%02x ", c)
			res = append(res, s...)
		} else {
			res = append(res, c)
		}
	}
	return res, true
}

// if d is a valid json, pretty-print it
// only used for debugging
func maybePrettyPrintJSON(d []byte) []byte {
	if d2, ok := asHex(d); ok {
		return d2
	}
	var m map[string]interface{}
	err := json.Unmarshal(d, &m)
	if err != nil {
		return d
	}
	d2, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return d
	}
	return d2
}

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
