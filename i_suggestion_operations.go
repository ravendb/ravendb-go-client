package ravendb

type ISuggestionOperations interface {
	WithOptions(options *SuggestionOptions) ISuggestionOperations
}
