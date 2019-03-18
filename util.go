package ravendb

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	// if true, does verbose logging.
	LogVerbose = false
)

func dbg(format string, args ...interface{}) {
	if LogVerbose {
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

func panicIfErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func i64toa(n int64) string {
	return strconv.FormatInt(n, 10)
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

func deepCopy(v interface{}) interface{} {
	// TODO: implement me
	return v
}

func builderWriteInt(b *strings.Builder, n int) {
	b.WriteString(strconv.Itoa(n))
}

func builderWriteFloat64(b *strings.Builder, f float64) {
	b.WriteString(fmt.Sprintf("%f", f))
}
