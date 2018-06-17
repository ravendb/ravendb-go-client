package ravendb

var (
	RQL_KEYWORDS = map[string]struct{}{
		"as":      struct{}{},
		"select":  struct{}{},
		"where":   struct{}{},
		"load":    struct{}{},
		"group":   struct{}{},
		"order":   struct{}{},
		"include": struct{}{},
	}
)

// QueryToken describes interface for query token
// In Java QueryToken is a base class that defines virtual writeTo and provides
// writeField. We make writeField a stand-alone helper function and make QueryToken
// an interface
type QueryToken interface {
	writeTo(*StringBuilder)
}

func QueryToken_writeField(writer *StringBuilder, field string) {
	_, keyWord := RQL_KEYWORDS[field]
	if keyWord {
		writer.append("'")
	}
	writer.append(field)

	if keyWord {
		writer.append("'")
	}
}
