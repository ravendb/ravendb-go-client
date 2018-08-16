package tests

type Employee struct {
	ID        string
	FirstName string
	LastName  string
}

func NewEmployee() *Employee {
	return &Employee{}
}

func (e *Employee) getId() string {
	return e.ID
}

func (e *Employee) setId(id string) {
	e.ID = id
}

func (e *Employee) getFirstName() string {
	return e.FirstName
}

func (e *Employee) setFirstName(firstName string) {
	e.FirstName = firstName
}

func (e *Employee) getLastName() string {
	return e.LastName
}

func (e *Employee) setLastName(lastName string) {
	e.LastName = lastName
}
