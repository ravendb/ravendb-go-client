package ravendb

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

// TODO: remove it, it only exists to make initial porting faster
type Object = interface{}

// TODO: remove it, it only exists to make initial porting faster
type String = string

// JsonNode represents JSON value
// equivalent of com.fasterxml.jackson.databind.JsonNode
type JsonNode = interface{}

// JsonNodeType represents a type of JSON value e.g. object, array.
// Equivalent of com.fasterxml.jackson.databind.node.JsonNodeType
// TODO: change to reflect.Type
type JsonNodeType = interface{}

// ObjectNode represents parsed JSON document in memory
// equivalent of com.fasterxml.jackson.databind.node.ObjectNode
type ObjectNode = map[string]interface{}

// ArrayNode represents result of BatchCommand, which is array of JSON objects
// it's a type alias so that it doesn't need casting when json marshalling
// equivalent of com.fasterxml.jackson.databind.node.ArrayNode
type ArrayNode = []ObjectNode

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

func StringUtils_isNotEmpty(s string) bool {
	return s != ""
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

func defaultTransformPlural(name string) string {
	return pluralize(name)
}

//https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/data/document_conventions.py#L45
func defaultTransformTypeTagName(name string) string {
	name = strings.ToLower(name)
	return defaultTransformPlural(name)
}

func min(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

func firstNonNilString(s1, s2 *string) *string {
	if s1 != nil {
		return s1
	}
	return s2
}

func firstNonEmptyString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

func firstNonZero(i1, i2 int) int {
	if i1 != 0 {
		return i1
	}
	return i2
}

func fieldNames(js ObjectNode) []string {
	var res []string
	for k := range js {
		res = append(res, k)
	}
	return res
}

// return a1 - a2
func stringArraySubtract(a1, a2 []string) []string {
	if len(a2) == 0 {
		return a1
	}
	if len(a1) == 0 {
		return nil
	}
	diff := make(map[string]struct{})
	for _, k := range a1 {
		diff[k] = struct{}{}
	}
	for _, k := range a2 {
		delete(diff, k)
	}
	if len(diff) == 0 {
		return nil
	}
	// TODO: pre-allocate
	var res []string
	for k := range diff {
		res = append(res, k)
	}
	return res
}

func fileExists(path string) bool {
	st, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return !st.IsDir()
}

func deepCopy(v interface{}) interface{} {
	// TOOD: implement me
	return v
}

func intArrayHasDuplicates(a []int) bool {
	if len(a) == 0 {
		return false
	}
	sort.Ints(a)
	prev := a[0]
	a = a[1:]
	for _, el := range a {
		if el == prev {
			return true
		}
		prev = el
	}
	return false
}

func intArrayContains(a []int, n int) bool {
	for _, el := range a {
		if el == n {
			return true
		}
	}
	return false
}
