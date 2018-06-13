package ravendb

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

func NewCompany() *Company {
	return &Company{}
}

func (c *Company) getAccountsReceivable() float64 {
	return c.AccountsReceivable
}

func (c *Company) setAccountsReceivable(accountsReceivable float64) {
	c.AccountsReceivable = accountsReceivable
}

func (c *Company) getId() string {
	return c.ID
}

func (c *Company) setId(id string) {
	c.ID = id
}

func (c *Company) getName() string {
	return c.Name
}

func (c *Company) setName(name string) {
	c.Name = name
}

func (c *Company) getDesc() string {
	return c.Desc
}

func (c *Company) setDesc(desc string) {
	c.Desc = desc
}

func (c *Company) getEmail() string {
	return c.Email
}

func (c *Company) setEmail(email string) {
	c.Email = email
}

func (c *Company) getAddress1() string {
	return c.Address1
}

func (c *Company) setAddress1(address1 string) {
	c.Address1 = address1
}

func (c *Company) getAddress2() string {
	return c.Address2
}

func (c *Company) setAddress2(address2 string) {
	c.Address2 = address2
}

func (c *Company) getAddress3() string {
	return c.Address3
}

func (c *Company) setAddress3(address3 string) {
	c.Address3 = address3
}

func (c *Company) getContacts() []*Contact {
	return c.Contacts
}

func (c *Company) setContacts(contacts []*Contact) {
	c.Contacts = contacts
}

func (c *Company) getPhone() int {
	return c.Phone
}

func (c *Company) setPhone(phone int) {
	c.Phone = phone
}

func (c *Company) getType() CompanyType {
	return c.Type
}

func (c *Company) setType(typ CompanyType) {
	c.Type = typ
}

func (c *Company) getEmployeesIds() []string {
	return c.EmployeesIds
}

func (c *Company) setEmployeesIds(employeesIds []string) {
	c.EmployeesIds = employeesIds
}
