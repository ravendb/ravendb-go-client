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

// getTypeName returns fully qualified (including package) name of the type,
// after traversing pointers.
// e.g. for struct Foo in main package, the type of Foo and *Foo is main.Foo
// TODO: test
func getTypeName(v interface{}) string {
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
