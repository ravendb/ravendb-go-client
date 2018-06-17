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

type QueryToken struct {
}

func NewQueryToken() *QueryToken {
	return &QueryToken{}
}

func (t *QueryToken) writeField(writer *StringBuilder, field string) {
	_, keyWord := RQL_KEYWORDS[field]
	if keyWord {
		writer.append("'")
	}
	writer.append(field)

	if keyWord {
		writer.append("'")
	}
}
