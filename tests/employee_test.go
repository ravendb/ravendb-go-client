package tests

// Employee represents an employee
type Employee struct {
	ID        string
	FirstName string
	LastName  string
}

func NewEmployee() *Employee {
	return &Employee{}
}
