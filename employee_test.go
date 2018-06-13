package ravendb

type Employee struct {
	ID        String
	FirstName String
	LastName  String
}

func NewEmployee() *Employee {
	return &Employee{}
}

func (e *Employee) getId() String {
	return e.ID
}

func (e *Employee) setId(id String) {
	e.ID = id
}

func (e *Employee) getFirstName() String {
	return e.FirstName
}

func (e *Employee) setFirstName(firstName String) {
	e.FirstName = firstName
}

func (e *Employee) getLastName() String {
	return e.LastName
}

func (e *Employee) setLastName(lastName String) {
	e.LastName = lastName
}
