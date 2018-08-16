package ravendb

var (
	SuggestionOptions_defaultOptions    = NewSuggestionOptions()
	SuggestionOptions_DEFAULT_ACCURACY  = float32(0.5)
	SuggestionOptions_DEFAULT_PAGE_SIZE = 15
	SuggestionOptions_DEFAULT_DISTANCE  = StringDistanceTypes_LEVENSHTEIN
	SuggestionOptions_DEFAULT_SORT_MODE = SuggestionSortMode_POPULARITY
)

type SuggestionOptions struct {
	pageSize int

	distance StringDistanceTypes

	accuracy float32

	sortMode SuggestionSortMode
}

func NewSuggestionOptions() *SuggestionOptions {
	return &SuggestionOptions{
		sortMode: SuggestionOptions_DEFAULT_SORT_MODE,
		distance: SuggestionOptions_DEFAULT_DISTANCE,
		accuracy: SuggestionOptions_DEFAULT_ACCURACY,
		pageSize: SuggestionOptions_DEFAULT_PAGE_SIZE,
	}
}

func (o *SuggestionOptions) getPageSize() int {
	return o.pageSize
}

func (o *SuggestionOptions) setPageSize(pageSize int) {
	o.pageSize = pageSize
}

func (o *SuggestionOptions) getDistance() StringDistanceTypes {
	return o.distance
}

func (o *SuggestionOptions) setDistance(distance StringDistanceTypes) {
	o.distance = distance
}

func (o *SuggestionOptions) getAccuracy() float32 {
	return o.accuracy
}

func (o *SuggestionOptions) setAccuracy(accuracy float32) {
	o.accuracy = accuracy
}

func (o *SuggestionOptions) getSortMode() SuggestionSortMode {
	return o.sortMode
}

func (o *SuggestionOptions) setSortMode(sortMode SuggestionSortMode) {
	o.sortMode = sortMode
}
