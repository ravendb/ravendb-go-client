package testing

type User struct{
	Id string `ravendb:"id"`
	Name, Lastname, AddressId string
	Count, Age int
}

type Product struct{
	Name string
}