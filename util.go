package ravendb

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	// if true, does verbose logging.
	// can be enabled by setting VERBOSE_LOG env variable to "true"
	LogVerbose = false

	// if true, logs summary of all HTTP requests i.e. "GET /foo" to stdout
	// can be enabled by setting LOG_HTTP_REQUEST_SUMMARY env variable to "true"
	LogRequestSummary = false

	// if true, logs request and response of failed http requests (i.e. those returning
	// status code >= 400) to stdout
	// can be enabled by setting LOG_FAILED_HTTP_REQUESTS env variable to "true"
	LogFailedRequests = false

	// if true, printing of failed reqeusts is delayed until PrintFailedRequests
	// is called.
	// can be enabled by setting LOG_FAILED_HTTP_REQUESTS_DELAYED env variable to "true"
	LogFailedRequestsDelayed = false

	// if true, logs all http requests/responses to a file for further inspection
	// this is for use in tests so the file has a fixed location:
	// logs/trace_${test_name}_go.txt
	// can be enabled by setting LOG_ALL_REQUESTS env variable to "true"
	LogAllRequests = false

	// if true, enables flaky tests
	// can be enabled by setting ENABLE_FLAKY_TESTS env variable to "true"
	EnableFlakyTests = false

	// if true, enable failing tests
	// can be enabled by setting ENABLE_FAILING_TESTS env variable to "true"
	EnableFailingTests = false

	// if true, we log RavenDB's output to stdout
	// can be enabled by setting LOG_RAVEN_SERVER env variable to "true"
	RavenServerVerbose = false
)

func SetStateFromEnv() {
	if !LogVerbose && isEnvVarTrue("VERBOSE_LOG") {
		LogVerbose = true
		fmt.Printf("Setting LogVerbose to true\n")
	}

	if !LogRequestSummary && isEnvVarTrue("LOG_HTTP_REQUEST_SUMMARY") {
		LogRequestSummary = true
		fmt.Printf("Setting LogRequestSummary to true\n")
	}

	if !LogFailedRequests && isEnvVarTrue("LOG_FAILED_HTTP_REQUESTS") {
		LogFailedRequests = true
		fmt.Printf("Setting LogFailedRequests to true\n")
	}

	if !LogFailedRequestsDelayed && isEnvVarTrue("LOG_FAILED_HTTP_REQUESTS_DELAYED") {
		LogFailedRequestsDelayed = true
		fmt.Printf("Setting LogFailedRequestsDelayed to true\n")
	}

	if !LogAllRequests && isEnvVarTrue("LOG_ALL_REQUESTS") {
		LogAllRequests = true
		fmt.Printf("Setting LogAllRequests to true\n")
	}

	if !RavenServerVerbose && isEnvVarTrue("LOG_RAVEN_SERVER") {
		RavenServerVerbose = true
		fmt.Printf("Setting RavenServerVerbose to true\n")
	}

	if !EnableFlakyTests && isEnvVarTrue("ENABLE_FLAKY_TESTS") {
		EnableFlakyTests = true
		fmt.Printf("Setting EnableFlakyTests to true\n")
	}

	if !EnableFailingTests && isEnvVarTrue("ENABLE_FAILING_TESTS") {
		EnableFailingTests = true
		fmt.Printf("Setting EnableFailingTests to true\n")
	}
}

func isEnvVarTrue(name string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch v {
	case "yes", "true":
		return true
	}
	return false
}

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

/*
// TODO:
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
*/

// TODO: remove
/*
// TODO: implement me exactly
func quoteKey2(s string, reservedSlash bool) string {
	// https://golang.org/src/net/url/url.go?s=7512:7544#L265
	return url.PathEscape(s)
}

func quoteKey(s string) string {
	return quoteKey2(s, false)
}
*/

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

// TODO: maybe sort or provide fieldNamesSorted when stable order matters
func FieldNames(js ObjectNode) []string {
	var res []string
	for k := range js {
		res = append(res, k)
	}
	return res
}

func FileExists(path string) bool {
	st, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return !st.IsDir()
}

func deepCopy(v interface{}) interface{} {
	// TODO: implement me
	return v
}

func InterfaceArrayContains(a []interface{}, v interface{}) bool {
	for _, el := range a {
		if el == v {
			return true
		}
	}
	return false
}

func builderWriteInt(b *strings.Builder, n int) {
	b.WriteString(strconv.Itoa(n))
}

func builderWriteFloat64(b *strings.Builder, f float64) {
	b.WriteString(fmt.Sprintf("%f", f))
}
