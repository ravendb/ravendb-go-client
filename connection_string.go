package ravendb

type ConnectionString struct {
	name string
	// Note: Java has this as a virtual function getType()
	typ ConnectionStringType
}

func (s *ConnectionString) GetName() string {
	return s.name
}

func (s *ConnectionString) setName(name string) {
	s.name = name
}

func (s *ConnectionString) getType() ConnectionStringType {
	return s.typ
}
