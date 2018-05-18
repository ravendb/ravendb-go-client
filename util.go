package ravendb

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func panicIf(cond bool, format string, args ...interface{}) {
	if cond {
		err := fmt.Errorf(format, args...)
		must(err)
	}
}

func isValidDbNameChar(c rune) bool {
	if c >= 'A' && c <= 'Z' {
		return true
	}
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	switch c {
	case '_', '-', '.':
		return true
	}
	return false
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/tools/utils.py#L47
// returns nil if db name is ok
func isDatabaseNameValid(dbName string) error {
	if dbName == "" {
		return errors.New("database name cannot be empty")
	}
	for _, c := range dbName {
		if !isValidDbNameChar(c) {
			return fmt.Errorf(`Database name can only contain only A-Z, a-z, _, . or - but was: %s`, dbName)
		}
	}
	return nil
}

/*
def quote_key(key, reserved_slash=False):
	reserved = '%:=&?~#+!$,;\'*[]'
	if reserved_slash:
		reserved += '/'
	return urllib.parse.quote(key, safe=reserved)
*/
// TODO: implement me exactly
func quoteKey2(s string, reservedSlash bool) string {
	// https://golang.org/src/net/url/url.go?s=7512:7544#L265
	return url.PathEscape(s)
}

func quoteKey(s string) string {
	return quoteKey2(s, false)
}

func quoteKeyWithSlash(s string) string {
	return quoteKey2(s, true)
}

// getFullTypeName returns fully qualified (including package) name of the type,
// after traversing pointers.
// e.g. for struct Foo in main package, the type of Foo and *Foo is main.Foo
func getFullTypeName(v interface{}) string {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typ := rv.Type()
	return typ.String()
}

// getShortTypeName returns a short (not including package) name of the type,
// after traversing pointers.
// e.g. for struct Foo, the type of Foo and *Foo is "Foo"
// This is equivalent to Python's v.__class__.__name__
func getShortTypeName(v interface{}) string {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typ := rv.Type()
	return typ.Name()
}

// TODO: make it more efficient by modifying the array in-place
func removeStringFromArray(pa *[]string, s string) {
	var res []string
	for _, s2 := range *pa {
		if s2 == s {
			continue
		}
		res = append(res, s2)
	}
	*pa = res
}

func stringArrayCopy(a []string) []string {
	n := len(a)
	if n == 0 {
		return nil
	}
	res := make([]string, n, n)
	for i := 0; i < n; i++ {
		res[i] = a[i]
	}
	return res
}

// delete "id" key from JSON object
// TODO: maybe should only
func deleteID(m map[string]interface{}) {
	for k := range m {
		if len(k) == 2 && strings.EqualFold(k, "id") {
			delete(m, k)
			return
		}
	}
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

func defaultTransformPlural(name string) string {
	return pluralize(name)
}

//https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/data/document_conventions.py#L45
func defaultTransformTypeTagName(name string) string {
	name = strings.ToLower(name)
	return defaultTransformPlural(name)
}

// TODO: move to DocumentConventions
func buildDefaultMetadata(entity interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if entity == nil {
		return res
	}
	fullTypeName := getFullTypeName(entity)
	typeName := getShortTypeName(entity)
	collectionName := defaultTransformPlural(typeName)
	res["@collection"] = collectionName
	res["Raven-Go-Type"] = fullTypeName
	return res
}

func getStructTypeOfReflectValue(rv reflect.Value) (reflect.Type, bool) {
	if rv.Type().Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typ := rv.Type()
	if typ.Kind() == reflect.Struct {
		return typ, true
	}
	return typ, false
}

func getStructTypeOfValue(v interface{}) (reflect.Type, bool) {
	rv := reflect.ValueOf(v)
	return getStructTypeOfReflectValue(rv)
}

// given a json represented as map and type of a struct
func makeStructFromJSONMap(typ reflect.Type, js ObjectNode) interface{} {
	panicIf(typ.Kind() != reflect.Struct, "rv should be of type Struct but is %s", typ.String())
	rvNew := reflect.New(typ)
	d, err := json.Marshal(js)
	must(err)
	v := rvNew.Interface()
	err = json.Unmarshal(d, v)
	must(err)
	return v
}

func convertToEntity(entityType reflect.Type, id string, document ObjectNode) interface{} {
	res := makeStructFromJSONMap(entityType, document)
	// TODO: set id on res
	return res
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

func firstNonEmptyString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}
