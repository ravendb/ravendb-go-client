package ravendb

var _ ISuggestionBuilder = &SuggestionBuilder{}
var _ ISuggestionOperations = &SuggestionBuilder{}

type SuggestionBuilder struct {
	_term  *SuggestionWithTerm
	_terms *SuggestionWithTerms
}

func NewSuggestionBuilder() *SuggestionBuilder {
	return &SuggestionBuilder{}
}

func (b *SuggestionBuilder) ByField(fieldName string, term string, terms ...string) ISuggestionOperations {
	panicIf(fieldName == "", "fieldName cannot be empty")
	panicIf(term == "", "term cannot be empty")
	if len(terms) > 0 {
		b._term = NewSUggestionWithTerm(fieldName)
		b._term.Term = term
	} else {
		b._terms = NewSUggestionWithTerms(fieldName)
		b._terms.Terms = append([]string{term}, terms...)
	}
	return b
}

func (b *SuggestionBuilder) getSuggestion() SuggestionBase {
	if b._term != nil {
		return b._term
	}

	return b._terms
}

func (b *SuggestionBuilder) WithOptions(options *SuggestionOptions) ISuggestionOperations {
	b.getSuggestion().SetOptions(options)

	return b
}
