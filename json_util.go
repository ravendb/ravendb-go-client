package ravendb

import (
	"encoding/json"
	"strconv"
	"strings"
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
