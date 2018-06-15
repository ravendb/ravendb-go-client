package ravendb

import (
	"encoding/json"
	"strconv"
)

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

func jsonGetAsString(doc ObjectNode, key string) string {
	return jsonGetAsText(doc, key)
}

func jsonGetAsText(doc ObjectNode, key string) string {
	v, ok := doc[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func jsonGetAsInt(doc ObjectNode, key string) (int, bool) {
	v, ok := doc[key]
	if !ok {
		return 0, false
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
