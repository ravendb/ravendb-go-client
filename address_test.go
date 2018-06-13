package ravendb

type Address struct {
	ID      string `json:"ID"`
	Country string `json:"Country"`
	City    string `json:"City"`
	Street  string `json:"Street"`
	ZipCode int    `json:"ZipCode"`
}

func NewAddress() *Address {
	return &Address{}
}

func (a *Address) getId() string {
	return a.ID
}

func (a *Address) setId(id string) {
	a.ID = id
}

func (a *Address) getCountry() string {
	return a.Country
}

func (a *Address) setCountry(country string) {
	a.Country = country
}

func (a *Address) getCity() string {
	return a.City
}

func (a *Address) setCity(city string) {
	a.City = city
}

func (a *Address) getStreet() string {
	return a.Street
}

func (a *Address) setStreet(street string) {
	a.Street = street
}

func (a *Address) getZipCode() int {
	return a.ZipCode
}

func (a *Address) setZipCode(zipCode int) {
	a.ZipCode = zipCode
}
