package ravendb

// SuggestionsBuilder helps to build argument to SuggestUsing
type SuggestionBuilder struct {
	term  *SuggestionWithTerm
	terms *SuggestionWithTerms
}

func NewSuggestionBuilder() *SuggestionBuilder {
	return &SuggestionBuilder{}
}

func (b *SuggestionBuilder) ByField(fieldName string, term string, terms ...string) *SuggestionBuilder {
	panicIf(fieldName == "", "fieldName cannot be empty")
	panicIf(term == "", "term cannot be empty")
	if len(terms) > 0 {
		b.terms = NewSuggestionWithTerms(fieldName)
		b.terms.Terms = append([]string{term}, terms...)
	} else {
		b.term = NewSuggestionWithTerm(fieldName)
		b.term.Term = term
	}
	return b
}

func (b *SuggestionBuilder) GetSuggestion() SuggestionBase {
	if b.term != nil {
		return b.term
	}

	return b.terms
}

func (b *SuggestionBuilder) WithOptions(options *SuggestionOptions) *SuggestionBuilder {
	b.GetSuggestion().SetOptions(options)

	return b
}
