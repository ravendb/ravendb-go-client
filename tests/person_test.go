package tests

type Person struct {
	ID        string
	Name      string
	AddressId string
}

func (p *Person) getId() string {
	return p.ID
}

func (p *Person) setId(id string) {
	p.ID = id
}

func (p *Person) GetName() string {
	return p.Name
}

func (p *Person) setName(name string) {
	p.Name = name
}

func (p *Person) getAddressId() string {
	return p.AddressId
}

func (p *Person) setAddressId(addressId string) {
	p.AddressId = addressId
}
