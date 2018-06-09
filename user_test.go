package ravendb

// for tests only
type User struct {
	ID        String
	Name      String
	LastName  String
	AddressId String
	Count     int
	Age       int
}

func (u *User) getId() String {
	return u.ID
}

func (u *User) getName() String {
	return u.Name
}

func (u *User) getLastName() String {
	return u.LastName
}

func (u *User) getAddressId() String {
	return u.AddressId
}

func (u *User) getCount() int {
	return u.Count
}

func (u *User) getAge() int {
	return u.Age
}

func (u *User) setId(id String) {
	u.ID = id
}

func (u *User) setName(name String) {
	u.Name = name
}

func (u *User) setLastName(lastName String) {
	u.LastName = lastName
}

func (u *User) setAddressId(addressId String) {
	u.AddressId = addressId
}

func (u *User) setCount(count int) {
	u.Count = count
}

func (u *User) setAge(age int) {
	u.Age = age
}
