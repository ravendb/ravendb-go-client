package ravendb

// PatchResult describes server results of patch command
type PatchResult struct {
	Status           PatchStatus `json:"Status"`
	ModifiedDocument ObjectNode  `json:"ModifiedDocument"`
	OriginalDocument ObjectNode  `json:"OriginalDocument"`
	Debug            ObjectNode  `json:"Debug"`

	// TODO: can this ever be null? If not, use string for type
	ChangeVector *string `json:"ChangeVector"`
	Collection   string  `json:"Collection"`
}
