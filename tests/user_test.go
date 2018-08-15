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

func NewUser() *User {
	return &User{}
}

func (u *User) getId() string {
	return u.ID
}

func (u *User) GetName() *string {
	return u.Name
}

func (u *User) getLastName() *string {
	return u.LastName
}

func (u *User) getAddressId() *string {
	return u.AddressId
}

func (u *User) getCount() int {
	return u.Count
}

func (u *User) getAge() int {
	return u.Age
}

func (u *User) setId(id string) {
	u.ID = id
}

func (u *User) setName(name string) {
	u.Name = &name
}

func (u *User) setLastName(lastName string) {
	u.LastName = &lastName
}

func (u *User) setAddressId(addressId string) {
	u.AddressId = &addressId
}

func (u *User) setCount(count int) {
	u.Count = count
}

func (u *User) setAge(age int) {
	u.Age = age
}
