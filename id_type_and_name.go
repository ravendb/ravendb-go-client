package ravendb

type IdTypeAndName struct {
	id   string
	typ  CommandType
	name string
}

func NewIdTypeAndName(id string, typ CommandType, name string) IdTypeAndName {
	return IdTypeAndName{
		id:   id,
		typ:  typ,
		name: name,
	}
}

// TODO: remove setter and getter functions after porting most of the code
func (t *IdTypeAndName) getId() string {
	return t.id
}

func (t *IdTypeAndName) setId(id string) {
	t.id = id
}

func (t *IdTypeAndName) getType() CommandType {
	return t.typ
}

func (t *IdTypeAndName) setType(typ CommandType) {
	t.typ = typ
}

func (t *IdTypeAndName) getName() string {
	return t.name
}

func (t *IdTypeAndName) setName(name string) {
	t.name = name
}
