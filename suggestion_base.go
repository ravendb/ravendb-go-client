package ravendb

// for documentation purposes only
// TODO: could be simplified. We really only need SuggestionWithTerms, which is a superset
// of SuggestionWithTerm
type SuggestionBase interface {
	SetOptions(*SuggestionOptions)
}

type SuggestionCommon struct {
	Field   string
	Options *SuggestionOptions
}

func (s *SuggestionCommon) SetOptions(options *SuggestionOptions) {
	s.Options = options
}
