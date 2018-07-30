package ravendb

import (
	"bytes"
	"encoding/json"
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
	return jsonGetAsText(doc, key)
}

func jsonGetAsText(doc ObjectNode, key string) (string, bool) {
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
func structToJSONMap(v interface{}) map[string]interface{} {
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
func valueToTree(v interface{}) ObjectNode {
	return structToJSONMap(v)
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
