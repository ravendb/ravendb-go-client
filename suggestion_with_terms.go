package ravendb

var _ SuggestionBase = &SuggestionWithTerms{}

type SuggestionWithTerms struct {
	SuggestionCommon
	Terms []string
}

func NewSUggestionWithTerms(field string) *SuggestionWithTerms {
	res := &SuggestionWithTerms{}
	res.Field = field
	return res
}
