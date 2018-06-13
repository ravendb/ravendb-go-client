package ravendb

type Person struct {
	ID        String
	Name      String
	AddressId String
}

func (p *Person) getId() String {
	return p.ID
}

func (p *Person) setId(id String) {
	p.ID = id
}

func (p *Person) getName() String {
	return p.Name
}

func (p *Person) setName(name String) {
	p.Name = name
}

func (p *Person) getAddressId() String {
	return p.AddressId
}

func (p *Person) setAddressId(addressId String) {
	p.AddressId = addressId
}
