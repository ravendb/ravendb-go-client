package ravendb

// PatchResult describes server results of patch command
type PatchResult struct {
	Status           PatchStatus            `json:"Status"`
	ModifiedDocument map[string]interface{} `json:"ModifiedDocument"`
	OriginalDocument map[string]interface{} `json:"OriginalDocument"`
	Debug            map[string]interface{} `json:"Debug"`
	LastModified     Time                   `json:"LastModified"`

	// TODO: can this ever be null? If not, use string for type
	ChangeVector *string `json:"ChangeVector"`
	Collection   string  `json:"Collection"`
}
