package ravendb

type Contact struct {
	ID        String
	FirstName String
	Surname   String
	Email     String
}

func NewContact() *Contact {
	return &Contact{}
}

func (c *Contact) getId() String {
	return c.ID
}

func (c *Contact) setId(id String) {
	c.ID = id
}

func (c *Contact) getFirstName() String {
	return c.FirstName
}

func (c *Contact) setFirstName(firstName String) {
	c.FirstName = firstName
}

func (c *Contact) getSurname() String {
	return c.Surname
}

func (c *Contact) setSurname(surname String) {
	c.Surname = surname
}

func (c *Contact) getEmail() String {
	return c.Email
}

func (c *Contact) setEmail(email String) {
	c.Email = email
}
