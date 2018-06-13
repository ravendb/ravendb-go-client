package ravendb

type Contact struct {
	ID        string
	FirstName string
	Surname   string
	Email     string
}

func NewContact() *Contact {
	return &Contact{}
}

func (c *Contact) getId() string {
	return c.ID
}

func (c *Contact) setId(id string) {
	c.ID = id
}

func (c *Contact) getFirstName() string {
	return c.FirstName
}

func (c *Contact) setFirstName(firstName string) {
	c.FirstName = firstName
}

func (c *Contact) getSurname() string {
	return c.Surname
}

func (c *Contact) setSurname(surname string) {
	c.Surname = surname
}

func (c *Contact) getEmail() string {
	return c.Email
}

func (c *Contact) setEmail(email string) {
	c.Email = email
}
