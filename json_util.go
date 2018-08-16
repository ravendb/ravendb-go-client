package ravendb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// JsonNode represents JSON value
// equivalent of com.fasterxml.jackson.databind.JsonNode
type JsonNode = interface{}

// JsonNodeType represents a type of JSON value e.g. object, array.
// Equivalent of com.fasterxml.jackson.databind.node.JsonNodeType
// TODO: change to reflect.Type?
type JsonNodeType = interface{}

// ObjectNode represents parsed JSON document in memory
// equivalent of com.fasterxml.jackson.databind.node.ObjectNode
type ObjectNode = map[string]interface{}

// TreeNode is equivalent of com.fasterxml.jackson.databind.TreeNode
// in terms of Go's json package, it's the same as interface{} becuase
// interface{} combines both the value and its type
type TreeNode = interface{}

// ArrayNode represents result of BatchCommand, which is array of JSON objects
// it's a type alias so that it doesn't need casting when json marshalling
// equivalent of com.fasterxml.jackson.databind.node.ArrayNode
type ArrayNode = []ObjectNode

func jsonGetAsTextPointer(doc ObjectNode, key string) *string {
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

func jsonGetAsString(doc ObjectNode, key string) (string, bool) {
	return JsonGetAsText(doc, key)
}

func JsonGetAsText(doc ObjectNode, key string) (string, bool) {
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

func jsonGetAsInt(doc ObjectNode, key string) (int, bool) {
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

func jsonGetAsBool(doc ObjectNode, key string) (bool, bool) {
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
func StructToJSONMap(v interface{}) map[string]interface{} {
	d, err := json.Marshal(v)
	must(err)
	var res map[string]interface{}
	err = json.Unmarshal(d, &res)
	must(err)
	return res
}

// given a json in the form of map[string]interface{}, de-serialize it to a struct
// TODO: could be faster
func structFromJSONMap(js ObjectNode, v interface{}) error {
	d, err := json.Marshal(js)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, v)
}

// matches a Java naming from EnityMapper
func ValueToTree(v interface{}) ObjectNode {
	return StructToJSONMap(v)
}

// copyJSONMap makes a deep copy of map[string]interface{}
// TODO: possibly not the fastest way to do it
func copyJSONMap(v map[string]interface{}) map[string]interface{} {
	d, err := json.Marshal(v)
	must(err)
	var res map[string]interface{}
	err = json.Unmarshal(d, &res)
	must(err)
	return res
}

// jsonDecodeFirst decode first JSON object from d
// This is like json.Unmarshal() but allows for d
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

func decodeJSONFromReader(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
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
