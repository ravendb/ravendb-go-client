package ravendb

import "strings"

func isRqlTokenKeyword(s string) bool {
	switch s {
	case "as", "select", "where", "load",
		"group", "order", "include":
		return true
	}
	return false
}

// QueryToken describes interface for query token
// In Java QueryToken is a base class that defines virtual writeTo and provides
// writeField. We make writeField a stand-alone helper function and make QueryToken
// an interface
type QueryToken interface {
	WriteTo(*strings.Builder)
}

func QueryToken_writeField(writer *strings.Builder, field string) {
	isKeyWord := isRqlTokenKeyword(field)
	if isKeyWord {
		writer.WriteString("'")
	}
	writer.WriteString(field)

	if isKeyWord {
		writer.WriteString("'")
	}
}
