package tests

type Address struct {
	ID      string
	Country string `json:"country"`
	City    string `json:"city"`
	Street  string `json:"street"`
	ZipCode int    `json:"zipCode"`
}
