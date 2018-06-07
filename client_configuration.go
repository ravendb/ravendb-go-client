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

func (c *ClientConfiguration) getEtag() int {
	return c.Etag
}

func (c *ClientConfiguration) setEtag(etag int) {
	c.Etag = etag
}

func (c *ClientConfiguration) isDisabled() bool {
	return c.Disabled
}

func (c *ClientConfiguration) setDisabled(disabled bool) {
	c.Disabled = disabled
}

func (c *ClientConfiguration) getMaxNumberOfRequestsPerSession() int {
	return c.MaxNumberOfRequestsPerSession
}

func (c *ClientConfiguration) setMaxNumberOfRequestsPerSession(maxNumberOfRequestsPerSession int) {
	c.MaxNumberOfRequestsPerSession = maxNumberOfRequestsPerSession
}

func (c *ClientConfiguration) getReadBalanceBehavior() ReadBalanceBehavior {
	return c.ReadBalanceBehavior
}

func (c *ClientConfiguration) setReadBalanceBehavior(readBalanceBehavior ReadBalanceBehavior) {
	c.ReadBalanceBehavior = readBalanceBehavior
}
