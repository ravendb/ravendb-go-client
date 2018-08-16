package ravendb

type ClientConfiguration struct {
	Etag     int  `json:"Etag"`
	Disabled bool `json:"Disabled"`
	// TODO: should this be *int ?
	MaxNumberOfRequestsPerSession int                 `json:"MaxNumberOfRequestsPerSession"`
	ReadBalanceBehavior           ReadBalanceBehavior `json:"ReadBalanceBehavior"`
}

func NewClientConfiguration() *ClientConfiguration {
	return &ClientConfiguration{}
}

func (c *ClientConfiguration) GetEtag() int {
	return c.Etag
}

func (c *ClientConfiguration) SetEtag(etag int) {
	c.Etag = etag
}

func (c *ClientConfiguration) IsDisabled() bool {
	return c.Disabled
}

func (c *ClientConfiguration) SetDisabled(disabled bool) {
	c.Disabled = disabled
}

func (c *ClientConfiguration) GetMaxNumberOfRequestsPerSession() int {
	return c.MaxNumberOfRequestsPerSession
}

func (c *ClientConfiguration) SetMaxNumberOfRequestsPerSession(maxNumberOfRequestsPerSession int) {
	c.MaxNumberOfRequestsPerSession = maxNumberOfRequestsPerSession
}

func (c *ClientConfiguration) GetReadBalanceBehavior() ReadBalanceBehavior {
	return c.ReadBalanceBehavior
}

func (c *ClientConfiguration) SetReadBalanceBehavior(readBalanceBehavior ReadBalanceBehavior) {
	c.ReadBalanceBehavior = readBalanceBehavior
}
