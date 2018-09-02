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

// In Java QueryToken is a base class that defines virtual writeTo and provides
// writeField. We make writeField a stand-alone helper function and make queryToken
// an interface
type queryToken interface {
	writeTo(*strings.Builder)
}

func writeQueryTokenField(writer *strings.Builder, field string) {
	isKeyWord := isRqlTokenKeyword(field)
	if isKeyWord {
		writer.WriteString("'")
		writer.WriteString(field)
		writer.WriteString("'")
		return
	}

	writer.WriteString(field)
}
