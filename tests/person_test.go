package tests

// Person represents a person
type Person struct {
	ID        string
	Name      string
	AddressId string `json:"AddressId"`
}
