package ravendb

type ClientConfiguration struct {
	Etag       int  `json:"Etag"`
	IsDisabled bool `json:"Disabled"`
	// TODO: should this be *int ?
	MaxNumberOfRequestsPerSession int                 `json:"MaxNumberOfRequestsPerSession"`
	ReadBalanceBehavior           ReadBalanceBehavior `json:"ReadBalanceBehavior"`
}
