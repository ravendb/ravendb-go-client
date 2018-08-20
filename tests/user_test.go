package tests

// for tests only
type User struct {
	ID        string
	Name      *string `json:"name"`
	LastName  *string `json:"lastName"`
	AddressId *string `json:"addressId"`
	Count     int     `json:"count"`
	Age       int     `json:"age"`
}

func (u *User) setName(name string) {
	u.Name = &name
}

func (u *User) setLastName(lastName string) {
	u.LastName = &lastName
}
