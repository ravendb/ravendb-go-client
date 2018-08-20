package ravendb

var (
	SuggestionOptions_defaultOptions    = NewSuggestionOptions()
	SuggestionOptions_DEFAULT_ACCURACY  = float32(0.5)
	SuggestionOptions_DEFAULT_PAGE_SIZE = 15
	SuggestionOptions_DEFAULT_DISTANCE  = StringDistanceTypes_LEVENSHTEIN
	SuggestionOptions_DEFAULT_SORT_MODE = SuggestionSortMode_POPULARITY
)

type SuggestionOptions struct {
	PageSize int

	Distance StringDistanceTypes

	Accuracy float32

	SortMode SuggestionSortMode
}

func NewSuggestionOptions() *SuggestionOptions {
	return &SuggestionOptions{
		SortMode: SuggestionOptions_DEFAULT_SORT_MODE,
		Distance: SuggestionOptions_DEFAULT_DISTANCE,
		Accuracy: SuggestionOptions_DEFAULT_ACCURACY,
		PageSize: SuggestionOptions_DEFAULT_PAGE_SIZE,
	}
}
