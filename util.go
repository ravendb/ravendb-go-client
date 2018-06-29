package ravendb

import (
	"fmt"
	"net/url"
	"os"
)

// TODO: remove it, it only exists to make initial porting faster
type Object = interface{}

// TODO: remove it, it only exists to make initial porting faster
type String = string

var (
	verboseLog = false
)

func dbg(format string, args ...interface{}) {
	if verboseLog {
		fmt.Printf(format, args...)
	}
}

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

// TODO: implement me exactly
func quoteKey2(s string, reservedSlash bool) string {
	// https://golang.org/src/net/url/url.go?s=7512:7544#L265
	return url.PathEscape(s)
}

func quoteKey(s string) string {
	return quoteKey2(s, false)
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
