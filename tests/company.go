package tests

type Company struct {
	AccountsReceivable float64
	ID                 string
	Name               string
	Desc               string
	Email              string
	Address1           string
	Address2           string
	Address3           string
	Contacts           []*Contact
	Phone              int
	Type               CompanyType
	EmployeesIds       []string
}
