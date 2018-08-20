package ravendb

var _ SuggestionBase = &SuggestionWithTerm{}

type SuggestionWithTerm struct {
	SuggestionCommon
	Term string
}

func NewSuggestionWithTerm(field string) *SuggestionWithTerm {
	res := &SuggestionWithTerm{}
	res.Field = field
	return res
}
