package ravendb

// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/util/Inflector.java#L14
// https://github.com/ixmatus/inflector
// https://github.com/tangzero/inflector : Go version but could be faster / simpler
// https://github.com/gedex/inflector : better Go version
// https://github.com/c9s/inflect : another one in Go

func pluralize(s string) string {
	// TODO: implement more sophisticated pluralization
	return s + "s"
}
