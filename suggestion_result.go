package ravendb

type SuggestionResult struct {
	Name        string   `json:"Name"`
	Suggestions []string `json:"Suggestions"`
}
