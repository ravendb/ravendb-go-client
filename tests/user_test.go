package tests

import "reflect"

var (
	userType = reflect.TypeOf(&User{})
)

type User struct {
	ID        string
	Name      *string `json:"name"`
	LastName  *string `json:"lastName"`
	AddressID string  `json:"addressId,omitempty"`
	Count     int     `json:"count"`
	Age       int     `json:"age"`
}

func (u *User) setName(name string) {
	u.Name = &name
}

func (u *User) setLastName(lastName string) {
	u.LastName = &lastName
}

func (u *User) setAge(age int) {
	u.Age = age
}
