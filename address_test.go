package ravendb

type Address struct {
	ID      String `json:"ID"`
	Country String `json:"Country"`
	City    String `json:"City"`
	Street  String `json:"Street"`
	ZipCode int    `json:"ZipCode"`
}

func NewAddress() *Address {
	return &Address{}
}

func (a *Address) getId() String {
	return a.ID
}

func (a *Address) setId(id String) {
	a.ID = id
}

func (a *Address) getCountry() String {
	return a.Country
}

func (a *Address) setCountry(country String) {
	a.Country = country
}

func (a *Address) getCity() String {
	return a.City
}

func (a *Address) setCity(city String) {
	a.City = city
}

func (a *Address) getStreet() String {
	return a.Street
}

func (a *Address) setStreet(street String) {
	a.Street = street
}

func (a *Address) getZipCode() int {
	return a.ZipCode
}

func (a *Address) setZipCode(zipCode int) {
	a.ZipCode = zipCode
}
