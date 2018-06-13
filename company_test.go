package ravendb

type Company struct {
	AccountsReceivable float64
	ID                 String
	Name               String
	Desc               String
	Email              String
	Address1           String
	Address2           String
	Address3           String
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

func (c *Company) getId() String {
	return c.ID
}

func (c *Company) setId(id String) {
	c.ID = id
}

func (c *Company) getName() String {
	return c.Name
}

func (c *Company) setName(name String) {
	c.Name = name
}

func (c *Company) getDesc() String {
	return c.Desc
}

func (c *Company) setDesc(desc String) {
	c.Desc = desc
}

func (c *Company) getEmail() String {
	return c.Email
}

func (c *Company) setEmail(email String) {
	c.Email = email
}

func (c *Company) getAddress1() String {
	return c.Address1
}

func (c *Company) setAddress1(address1 String) {
	c.Address1 = address1
}

func (c *Company) getAddress2() String {
	return c.Address2
}

func (c *Company) setAddress2(address2 String) {
	c.Address2 = address2
}

func (c *Company) getAddress3() String {
	return c.Address3
}

func (c *Company) setAddress3(address3 String) {
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
