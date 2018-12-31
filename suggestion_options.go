package ravendb

var (
	SuggestionOptionsDefaultOptions  = NewSuggestionOptions()
	SuggestionOptionsDefaultAccuracy = float32(0.5)
	SuggestionOptionsDefaultPageSize = 15
	SuggestionOptionsDefaultDistance = StringDistanceLevenshtein
	SuggestionOptionsDefaultSortMode = SuggestionSortModePopularity
)

type SuggestionOptions struct {
	PageSize int

	Distance StringDistanceTypes

	Accuracy float32

	SortMode SuggestionSortMode
}

func NewSuggestionOptions() *SuggestionOptions {
	return &SuggestionOptions{
		SortMode: SuggestionOptionsDefaultSortMode,
		Distance: SuggestionOptionsDefaultDistance,
		Accuracy: SuggestionOptionsDefaultAccuracy,
		PageSize: SuggestionOptionsDefaultPageSize,
	}
}
