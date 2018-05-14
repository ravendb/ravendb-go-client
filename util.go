package ravendb

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
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
